package ultron

import (
	"github.com/qastub/ultron/utils"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	slaveRunner struct {
		id      string
		gClient UltronClient
		once    sync.Once
		sendCh  chan *Result
		*baseRunner
	}
)

var (
	slaveStart = make(chan struct{}, 1)

	// ResultStreamBufferSize slave->master
	ResultStreamBufferSize = 100
)

func newSlaveRunner() *slaveRunner {
	return &slaveRunner{id: uuid.NewV4().String(), baseRunner: newBaseRunner()}
}

func (sl *slaveRunner) Connect(addr string, opts ...grpc.DialOption) {
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		Logger.Error("connect to MasterRunner failed", zap.Error(err))
		panic(err)
	}
	sl.gClient = NewUltronClient(conn)
}

func (sl *slaveRunner) handleMsg() {
	c, err := sl.gClient.Subscribe(context.Background(), &ClientInfo{Id: sl.id})
	if err != nil {
		c.CloseSend()
		sl.gClient.(*ultronClient).cc.Close()
		os.Exit(1)
	}
	Logger.Info("subscribed")

	for {
		msg, err := c.Recv()
		if err != nil {
			c.CloseSend()
			Logger.Error("got error", zap.Error(err))
			sl.gClient.(*ultronClient).cc.Close()
			os.Exit(1)
		}

		switch msg.Type {
		case Message_Disconnect:
			Logger.Warn("received message to disconnect")
			c.CloseSend()
			sl.gClient.(*ultronClient).cc.Close()
			os.Exit(0)

		case Message_RefreshConfig:
			conf := new(RunnerConfig)
			err = json.Unmarshal(msg.Data, conf)
			if err != nil {
				Logger.Error("unmarshal RunnerConfig failed", zap.Error(err))
			} else {
				Logger.Info("refreshed runner config", zap.Any("new RunnerConfig", conf))
				sl.WithConfig(conf)
			}

		case Message_StartAttack:
			Logger.Info("received message to start attack")
			if sl.GetStatus() == StatusBusy {
				Logger.Warn("SlaveRunner is running, ignore this message")
			} else {
				slaveStart <- struct{}{}
			}

		case Message_StopAttack:
			Logger.Info("reveived message to stop attack")
			if sl.GetStatus() == StatusBusy {
				StageRunnerStatusPipeline <- StatusStopped
			} else {
				Logger.Warn("SlaveRunner is not running, ignore this message")
			}

		case Message_Ping:
			Logger.Info(fmt.Sprintf("i am alive, SlaveRunner Status: %d", sl.GetStatus()))

		default:
			Logger.Warn("unknown message", zap.Any("received", msg))
		}
	}
}

func (sl *slaveRunner) handleResult() ResultHandleFunc {
	return func(r *Result) {
		sl.sendCh <- r
	}
}

func (sl *slaveRunner) sendStream(size int) {
	if sl.sendCh == nil {
		sl.sendCh = make(chan *Result, size)
	}

	stream, err := sl.gClient.Send(context.Background())
	if err != nil {
		panic(err)
	}

	for r := range sl.sendCh {
		err = stream.Send(r)
		if err != nil {
			Logger.Error("occur error on sending result to master", zap.Error(err))
			os.Exit(1)
			break
		}
	}
	stream.CloseSend()
}

//func (sl *slaveRunner) Start() {
//	if sl.gClient == nil {
//		panic("you should invoke Connect(addr string) method first")
//	}
//
//	if sl.task == nil {
//		panic("no task provided")
//	}
//
//	sl.once.Do(func() {
//		go sl.handleMsg()
//		go sl.sendStream(ResultStreamBufferSize)
//		slaveResultPipeline = newResultPipeline(SlaveResultPipelineBufferSize)
//		SlaveEventHook.AddResultHandleFunc(sl.handleResult())
//		SlaveEventHook.listen(slaveResultPipeline, slaveReportPipeline)
//	})
//
//	for {
//		sl.status = StatusIdle
//		Logger.Info("slaver: " + sl.id + " is ready")
//		<-slaveStart
//		sl.status = StatusBusy
//		Logger.Info("attack !!!")
//
//		hatchWorkers(sl.baseRunner, slaveResultPipeline)
//		sl.wg.Wait()
//		Logger.Info("attack stopped")
//	}
//}


//TODO
//简单粗暴，需要优化
func (sl *slaveRunner) Start() {
	for {
		sl.status = StatusIdle
		Logger.Info("slaver: " + sl.id + " is ready")
		<-slaveStart
		sl.start()
	}
}

func (sl *slaveRunner) start() {

	if sl.gClient == nil {
		panic("you should invoke Connect(addr string) method first")
	}

	if err := checkRunner(sl.baseRunner); err != nil {
		panic(err)
	}

	sl.once.Do(func() {
		go sl.handleMsg()
		go sl.sendStream(ResultStreamBufferSize)
		slaveResultPipeline = newResultPipeline(SlaveResultPipelineBufferSize)
		SlaveEventHook.AddResultHandleFunc(sl.handleResult())
		SlaveEventHook.listen(slaveResultPipeline, slaveReportPipeline)

		// ctrl+c退出,输出信号
		//signalCh := make(chan os.Signal, 1)
		//signal.Notify(signalCh, os.Interrupt)
		//go func() {
		//	<-signalCh
		//	Logger.Error("capatured interrupt signal")
		//	printReportToConsole(lr.stats.report(true))
		//	os.Exit(1)
		//}()
	})

	go CountNumbers2Stop(CounterPipeline, &sl.Config.Requests)

	go statusControl(StageRunnerStatusPipeline)

	pctx, pcancel := createCancelFunc(sl.baseRunner, _parentCtx)

	//[]time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	timers := utils.NewTimers(sl.GetStageRunningTime())
	Logger.Info("start to attack")
	sl.status = StatusBusy

	for {
		select {
		case <- pctx.Done():
			//fmt.Println("feedTicker.Stop()")
			sl.baseRunner.Done()
			//localReportPipeline <- lr.stats.report(true)
			Logger.Info("stages have be done. STOP!")
			StageRunnerStatusPipeline <- StatusStopped
			return
		case cc := <-timers.C:
			if cc >= 0 && cc <= len(sl.baseRunner.Config.Stages) - 1 {
				scc := sl.baseRunner.Config.Stages[cc]
				Logger.Info("start ", zap.Int("task：", cc))

				func() {
					if scc.Concurrence == 0 {
						// do nothing
						Logger.Info("keep Concurrence")
					} else {
						hatchWorkersCancelable(pctx, sl.baseRunner, scc, localResultPipeline, CounterPipeline)
					}
				}()

			} else {
				Logger.Info("pcancel()")
				pcancel()
				StageRunnerStatusPipeline <- StatusStopped
			}
		}
	}
}



