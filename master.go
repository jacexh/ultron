package ultron

import (
	"bytes"
	"encoding/json"
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

	grpc "google.golang.org/grpc"
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
	MasterListenAddr = "0.0.0.0:9500"

	defaultSessionPool *sessionPool
	serverStart        = make(chan struct{}, 1)
	serverStop         = make(chan struct{}, 1)
	serverInterrpt     = make(chan struct{}, 1)
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
		masterResultPipline <- ret
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
		Logger.Info("listen on " + mr.Listener.Addr().String())
		mr.Listener = lis
	}

	mr.serv = grpc.NewServer()
	RegisterUltronServer(mr.serv, mr)

	go func() {
		defer func() { os.Exit(1) }()
		Logger.Error("grpc server down", zap.Error(mr.serv.Serve(mr.Listener)))
	}()

	for {
		Logger.Info("ready to attack")
		<-serverStart
		Logger.Info("attack")
		mr.status = statusBusy
		mr.counts = 0
		mr.deadline = time.Time{}

		mr.once.Do(func() {
			masterReportPipeline = newReportPipeline(MasterReportPipelineBufferSize)
			masterResultPipline = newResultPipeline(MasterResultPipelineBufferSize)
			MasterEventHook.AddResultHandleFunc(mr.count, mr.stats.log)
			go MasterEventHook.listen(masterResultPipline, masterReportPipeline)

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt)
			go func() {
				<-signalCh
				Logger.Error("capatured interrupt signal")
				serverInterrpt <- struct{}{}
			}()
		})

		go func() {
			t := time.NewTicker(time.Millisecond * 200)
			for range t.C {
				if mr.IsFinished() {
					serverStop <- struct{}{}
					break
				}
			}
		}()

		err := defaultSessionPool.sendConfigToSlaves(mr.Config)
		if err != nil {
			Logger.Error("occur error", zap.Error(err))
			os.Exit(1)
		}

		mr.stats.reset()
		defaultSessionPool.batchSendMessage(Message_StartAttack, nil)

		if mr.Config.Duration > ZeroDuration { // 开始设置deadline
			mr.deadline = time.Now().Add(mr.Config.Duration)
			if mr.Config.HatchRate > 0 && mr.Config.Concurrence > mr.Config.HatchRate {
				secs := mr.Config.Concurrence / mr.Config.HatchRate
				if mr.Config.Concurrence%mr.Config.HatchRate > 0 {
					secs++
				}
				mr.baseRunner.deadline = time.Now().Add(time.Second * time.Duration(secs))
			}
			Logger.Info("set deadline", zap.Time("deadline", mr.deadline))
		}

		select {
		case <-serverStop: // 压测结束信号
			Logger.Info("stop to attack")
			defaultSessionPool.batchSendMessage(Message_StopAttack, nil)

		case <-serverInterrpt:
			defaultSessionPool.batchSendMessage(Message_Disconnect, nil)
			printReportToConsole(mr.stats.report(true))
			os.Exit(1)
		}
	}
}

func (sp *sessionPool) sendConfigToSlaves(rc *RunnerConfig) error {
	var e error
	cs := rc.split(sp.getSlaveCounts())
	index := 0
	sp.pool.Range(func(key, value interface{}) bool {
		c := cs[index]
		data, err := json.Marshal(c)
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
