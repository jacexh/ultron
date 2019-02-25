package ultron

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	masterRunner struct {
		addr     string
		Listener net.Listener
		serv     *grpc.Server
		once     sync.Once
		stats    *summaryStats
		*baseRunner
	}

	session struct {
		ch chan *Message
	}

	sessionPool struct {
		pool sync.Map
	}
)

var (
	// MasterRunner 分布式执行，主控执行入口
	MasterRunner *masterRunner

	// MasterListenAddr server端默认监听地址
	MasterListenAddr = ":9500"

	defaultSessionPool = &sessionPool{}
	// ServerStart MasterRunner启动压测的信号
	ServerStart = make(chan struct{}, 1)
	// ServerStop MasterRunner停止压测的信号
	ServerStop = make(chan struct{}, 1)
	// ServerInterrupt MasterRunner压测被中断的信号
	ServerInterrupt = make(chan struct{}, 1)

	pingInterval = time.Second * 15
)

func newSession() *session {
	return &session{
		ch: make(chan *Message),
	}
}

func newMasterRunner(addr string, ss *summaryStats) *masterRunner {
	return &masterRunner{
		addr:       addr,
		stats:      ss,
		baseRunner: newBaseRunner(),
	}
}

func (mr *masterRunner) Send(stream Ultron_SendServer) error {
	for {
		ret, err := stream.Recv()
		if err == io.EOF {
			stream.SendAndClose(&Ack{})
			Logger.Info("result stream closed")
			break
		}
		if err != nil {
			Logger.Error("occur error on receiving result from client")
			return err
		}
		masterResultPipeline <- ret
	}
	return nil
}

func (mr *masterRunner) Subscribe(c *ClientInfo, stream Ultron_SubscribeServer) error {
	if c.Id == "" {
		return errors.New("bad slaver id")
	}
	val, loaded := defaultSessionPool.pool.LoadOrStore(c.Id, newSession())
	if loaded {
		return errors.New("duplicated client id: " + c.Id)
	}

	for msg := range val.(*session).ch {
		err := stream.Send(msg)
		if err != nil || msg.Type == Message_Disconnect {
			defaultSessionPool.pool.Delete(c.Id)
			Logger.Warn("remove slaver: " + c.Id)
			if err != nil {
				Logger.Info("occur error on sending message", zap.Error(err))
				return err
			}
			break
		}
	}
	return io.EOF
}

func (mr *masterRunner) count(ret *Result) {
	atomic.AddUint64(&mr.counts, 1)
}

func (mr *masterRunner) Start() {
	if mr.Listener == nil {
		lis, err := net.Listen("tcp", mr.addr)
		if err != nil {
			Logger.Error(fmt.Sprintf("listen to %s failed", mr.addr), zap.Error(err))
			os.Exit(1)
		}
		mr.Listener = lis
	}

	Logger.Info("listen on " + mr.Listener.Addr().String())
	mr.serv = grpc.NewServer()
	RegisterUltronServer(mr.serv, mr)

	go func() {
		defer func() { os.Exit(1) }()
		Logger.Error("grpc server down", zap.Error(mr.serv.Serve(mr.Listener)))
	}()

	go func() {
		for {
			time.Sleep(pingInterval)
			defaultSessionPool.ping("")
		}
	}()

	mr.once.Do(func() {
		masterReportPipeline = newReportPipeline(MasterReportPipelineBufferSize)
		masterResultPipeline = newResultPipeline(MasterResultPipelineBufferSize)
		MasterEventHook.AddResultHandleFunc(mr.count, mr.stats.log)
		go MasterEventHook.listen(masterResultPipeline, masterReportPipeline)

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
		go func() {
			<-signalCh
			Logger.Error("capatured interrupt signal")
			ServerInterrupt <- struct{}{}
		}()
	})

	for {
		Logger.Info("ready to attack")

		select {
		case <-ServerInterrupt:
			Logger.Info("no runner was running, exit")
			os.Exit(0)
		case <-ServerStart:
		}

		err := mr.Config.check()
		if err != nil {
			Logger.Error("bad RunnerConfig", zap.Error(err))
			continue
		}

		Logger.Info("attack")
		mr.status = StatusBusy
		mr.counts = 0
		//mr.Deadline = time.Time{}

		go func() {
			t := time.NewTicker(time.Millisecond * 200)
			for range t.C {
				if isFinished(mr.baseRunner) {
					ServerStop <- struct{}{}
					break
				}
				if defaultSessionPool.getSlaveCounts() == 0 { // 如果slave都不存在了，测试就没必要进行下去
					Logger.Warn("all slavers was dead")
					mr.Done()
					ServerStop <- struct{}{}
					break
				}
			}
		}()

		go func() {
			t := time.NewTicker(StatsReportInterval)
			for range t.C {
				//if isFinished(mr.baseRunner) {
				//	break
				//}
				masterReportPipeline <- mr.stats.report(false)
			}
		}()

 		err = defaultSessionPool.sendConfigToSlaves(mr.baseRunner)
		if err != nil {
			Logger.Error("occur error", zap.Error(err))
			os.Exit(1)
		}

		mr.stats.reset()
		defaultSessionPool.batchSendMessage(Message_StartAttack, nil)

 		//TODO
		if mr.Config.Duration > ZeroDuration { // 开始设置deadline
			if mr.Config.HatchRate > 0 && mr.Config.Concurrence > mr.Config.HatchRate {
				secs := mr.Config.Concurrence / mr.Config.HatchRate
				if mr.Config.Concurrence%mr.Config.HatchRate > 0 {
					secs++
				}
				//TODO have a bug
				//mr.baseRunner.Deadline = time.Now().Add(time.Second * time.Duration(secs))
			}
		}

		select {
		case <-ServerStop: // 压测结束信号
			Logger.Info("stop to attack")
			defaultSessionPool.batchSendMessage(Message_StopAttack, nil)
			mr.baseRunner.Done()
			mr.stats.reset()

		case <-ServerInterrupt:
			defaultSessionPool.batchSendMessage(Message_Disconnect, nil)
			printReportToConsole(mr.stats.report(true))
			os.Exit(1)
		}
	}
}

func (sp *sessionPool) sendConfigToSlaves(br *baseRunner) error {
	var e error
	cs := br.Config.split(sp.getSlaveCounts())
	index := 0
	sp.pool.Range(func(key, value interface{}) bool {
		c := cs[index]
		br.WithConfig(c)

		data, err := json.Marshal(br)
		if err != nil {
			e = err
			return false
		}
		value.(*session).ch <- &Message{Type: Message_RefreshConfig, Data: data}
		index++
		return true
	})
	return e
}

func (sp *sessionPool) batchSendMessage(t Message_Type, d []byte) {
	sp.pool.Range(func(key, value interface{}) bool {
		msg := &Message{Type: t}
		if d != nil && bytes.Compare(d, []byte{}) > 0 {
			msg.Data = d
		}
		value.(*session).ch <- msg
		return true
	})
}

func (sp *sessionPool) getSlaveCounts() int {
	n := 0
	sp.pool.Range(func(key, value interface{}) bool {
		n++
		return true
	})
	return n
}

func (sp *sessionPool) ping(c string) {
	if c == "" {
		sp.batchSendMessage(Message_Ping, nil)
		return
	}

	sp.pool.Range(func(key, value interface{}) bool {
		cid := key.(string)
		if cid == c {
			value.(*session).ch <- &Message{Type: Message_Ping}
			return false
		}
		return true
	})
}
