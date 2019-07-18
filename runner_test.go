package ultron

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"
)

func Test_newBaseRunner(t *testing.T) {
	baserunner := newBaseRunner()
	fmt.Println(baserunner)

	stageConfig1 := NewStage()
	stageConfig2 := NewStage()
	runnerConfig := NewRunnerConfig()
	runnerConfig.AppendStage(stageConfig1).AppendStage(stageConfig2)

	if err := runnerConfig.check(); err != nil {
		t.Error(err)
	}

	fmt.Println(runnerConfig)
	for _, st := range runnerConfig.Stages {
		fmt.Println(st)
	}
}

func TestBaseRunner_WithConfig(t *testing.T) {
	baserunner := newBaseRunner()

	runnerConfig := NewRunnerConfig()
	baserunner.WithConfig(runnerConfig)
	//fmt.Println(baserunner)

}

func TestTask_Add2(t *testing.T) {
	task := NewTask()
	task.Add(newAttacker("a"), 10)
	task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)
	if len(task.attackers) != 3 {
		t.Error("task.Add wrong")
	}
}

func TestTask_Del(t *testing.T) {
	task := NewTask()
	a_weight := rand.Intn(50)
	b_weight := rand.Intn(50)
	c_weight := rand.Intn(50)
	c_attack := newAttacker("c")
	task.Add(newAttacker("a"), a_weight)
	task.Add(newAttacker("b"), b_weight)
	task.Add(c_attack, c_weight)
	task.Del(c_attack)
	if task.totalWeight != a_weight+b_weight {
		t.Error("task.Del totalWeight wrong")
	}
	if task.attackers[c_attack] != 0 {
		t.Error("task.Del attackers wrong")
	}
}

func TestRunnerConfig_AppendStage(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageConfig1 := NewStage(5*time.Minute, 50, 10)
	stageConfig2 := NewStage(5*time.Minute, 100, 10)
	stageConfig3 := NewStage(5*time.Minute, 70, 10)
	runnerconfig.AppendStages(stageConfig1, stageConfig2).AppendStages(stageConfig3)

	fmt.Println(runnerconfig)
	if len(runnerconfig.Stages) != 3 {
		t.Error("runnerconfig.AppendStages wrong")
	}
}

//多个stage
func TestRunnerConfig_UpdateStageConfig(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageConfig1 := NewStage(5*time.Minute, 200, 10)
	stageConfig2 := NewStage(7*time.Minute, 100, 10)
	stageConfig3 := NewStage(2*time.Hour, 600, 100)
	runnerconfig.AppendStages(stageConfig1, stageConfig2, stageConfig3)
	runnerconfig.updateStageConfig()

	//initconcurrence := []int{50, 100, 70}
	duration := []time.Duration{5 * time.Minute, 7 * time.Minute, 2 * time.Hour}
	concurrenceResult := []int{200, -100, 500}
	hatchRateResult := []int{10, 10, 100}

	for _, scc := range runnerconfig.Stages {
		fmt.Println(scc.Concurrence)
	}

	for i, scc := range runnerconfig.stagesChanged {
		if scc.Concurrence != concurrenceResult[i] {
			t.Error("UpdateStageRunnerConfig Concurrence wrong")
		}
		if scc.HatchRate != hatchRateResult[i] {
			t.Error("UpdateStageRunnerConfig HatchRate wrong")
		}
		if scc.Duration != duration[i] {
			t.Error("UpdateStageRunnerConfig Duration wrong")
		}
	}
}

//单个stage
func TestRunnerConfig_UpdateStageConfig2(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageConfig1 := NewStage(5*time.Minute, 200, 11000)

	runnerconfig.AppendStages(stageConfig1)
	runnerconfig.updateStageConfig()

	//initconcurrence := []int{50, 100, 70}
	duration := []time.Duration{5 * time.Minute}
	concurrenceResult := []int{200}
	hatchRateResult := []int{11000}

	for _, scc := range runnerconfig.Stages {
		fmt.Println(scc.Concurrence)
	}

	for i, scc := range runnerconfig.Stages {
		if scc.Concurrence != concurrenceResult[i] {
			t.Error("UpdateStageRunnerConfig Concurrence wrong")
		}
		if scc.HatchRate != hatchRateResult[i] {
			t.Error("UpdateStageRunnerConfig HatchRate wrong")
		}
		if scc.Duration != duration[i] {
			t.Error("UpdateStageRunnerConfig Duration wrong")
		}
	}
}

func TestNewStageConfig(t *testing.T) {
	d := 10 * time.Hour
	c := rand.Intn(1000000)
	h := rand.Intn(1000000)
	StageConfig := NewStage(d, c, h)

	if StageConfig.Duration != d && StageConfig.Concurrence != c && StageConfig.HatchRate != h {
		t.Error("NewStage wrong")
	}
}

//stage
func TestRunnerConfig_Check(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageConfig1 := NewStage(5*time.Minute, 50, 2)
	stageConfig2 := NewStage(5*time.Minute, 100, 1)
	stageConfig3 := NewStage(5*time.Minute, 70, 6)
	runnerconfig.AppendStages(stageConfig1, stageConfig2, stageConfig3)

	fmt.Println(runnerconfig)
	if err := runnerconfig.check(); err != nil {
		t.Error("stagerunnerconfig.MinWait/MaxWait wrong")
	}
}

//增加cancelfunc
func TestBaseRunner_AddCancelFunc(t *testing.T) {

	BaseRunner := newBaseRunner()
	ctx, cancel1 := context.WithCancel(context.Background())
	_, cancel2 := context.WithDeadline(ctx, time.Now())

	BaseRunner.AddCancelFunc(&cancel1)
	BaseRunner.AddCancelFunc(&cancel2)
	fmt.Println(BaseRunner.cancels.cancels)

	//stageRunner.cancels = append(stageRunner.cancels, cancel)
	if len(BaseRunner.cancels.cancels) != 2 {
		t.Error("StageRunner_addCancelFunc wrong")
	}
}

//错误配置 raise err
func TestRunnerConfig_Check3(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageConfig1 := NewStage(5*time.Minute, 50, 2)
	stageConfig2 := NewStage(5*time.Minute, 100, 1)
	runnerconfig.AppendStages(stageConfig1, stageConfig2)

	runnerconfig.Concurrence = 1000
	runnerconfig.HatchRate = 12312317

	if err := runnerconfig.check(); err == nil {
		t.Error("同时定义v1runner及stage")
	}
}

func TestBaseRunner_CheckRunner(t *testing.T) {
	task := NewTask()

	runnerconfig := NewRunnerConfig()
	stageConfig1 := NewStage(5*time.Minute, 50, 2)
	stageConfig2 := NewStage(5*time.Minute, 100, 1)
	runnerconfig.AppendStages(stageConfig1, stageConfig2)

	runnerconfig.HatchRate = 134

	base := newBaseRunner()
	base.WithTask(task)
	base.Config = runnerconfig
	if err := checkRunner(base); err != nil {
		t.Error("checkRunner error ", err)
	}

	//不能同时配置Concurrence 及 stage
	runnerconfig.Concurrence = 1000
	base.Config = runnerconfig
	if err := checkRunner(base); err == nil {
		t.Error("checkRunner error ")
	}
}

//兼容v1
func TestRunnerConfig_Check2(t *testing.T) {
	d := 1 * time.Hour
	c := 100000
	h := 123124
	runnerconfig := NewRunnerConfig()
	runnerconfig.Duration = d
	runnerconfig.Concurrence = c
	runnerconfig.HatchRate = h

	if len(runnerconfig.Stages) != 0 {
		t.Error("before runnerconfig.Stages is not 0")
	}
	if runnerconfig.Concurrence != c {
		t.Error("before runnerconfig.Concurrence is not equal before Concurrence")
	}

	if err := runnerconfig.check(); err != nil {
		t.Error("runnerconfig.Duration wrong")
	}

	//校验update之后的数据
	if len(runnerconfig.Stages) != 1 {
		t.Error("after runnerconfig.Stages is not 0")
	}
	if runnerconfig.Concurrence != 0 {
		t.Error("after runnerconfig.Concurrence is not equal before Concurrence")
	}

	if runnerconfig.Stages[0].Concurrence != c {
		t.Error("after runnerconfig.Stages[0].Concurrence != 100000")
	}
	if runnerconfig.Stages[0].Duration != d {
		t.Error("after runnerconfig.Stages[0].Concurrence != 100000")
	}
	if runnerconfig.Stages[0].HatchRate != h {
		t.Error("after runnerconfig.Stages[0].Concurrence != 100000")
	}
}

func TestRunnerConfig_Check4(t *testing.T) {
	d := 1 * time.Hour
	c := 100000
	h := -10
	runnerconfig := NewRunnerConfig()
	runnerconfig.Duration = d
	runnerconfig.Concurrence = c
	runnerconfig.HatchRate = h

	if err := runnerconfig.check(); err != nil {
		t.Error("runnerconfig.HatchRate < 0 ")
	}
}

func TestRunnerConfig_Check5(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig := NewStage(10*time.Minute, 1000, 0)
	runnerconfig.AppendStages(stageconfig)

	if err := runnerconfig.check(); err != nil {
		t.Error("runnerconfig.HatchRate < 0 ")
	}
}

func TestRunnerConfig_Check6(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig := NewStage(10*time.Minute, 1000, -250)
	runnerconfig.AppendStages(stageconfig)

	if err := runnerconfig.check(); err == nil {
		t.Error("runnerconfig.HatchRate < 0 ")
	}
}

func TestRunnerConfig_Check7(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig := NewStage(10*time.Minute, -1000, 250)
	runnerconfig.AppendStages(stageconfig)

	if err := runnerconfig.check(); err == nil {
		t.Error("runnerconfig.Concurrence < 0 ")
	}
}

func TestRunnerConfig_Check8(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig := NewStage(10*time.Minute, 0, 250)
	runnerconfig.AppendStages(stageconfig)

	if err := runnerconfig.check(); err == nil {
		t.Error("runnerconfig.Concurrence < 0 ")
	}
}

//非最后一个stage设置时长为0，报错
func TestRunnerConfig_Check9(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig1 := NewStage(10*time.Minute, 100, 250)
	stageconfig2 := NewStage(0*time.Minute, 100, 250)
	stageconfig3 := NewStage(10*time.Minute, 10, 250)

	runnerconfig.AppendStages(stageconfig1, stageconfig2, stageconfig3)

	if err := runnerconfig.check(); err == nil {
		t.Error("runnerconfig.Concurrence < 0 ")
	}
}

//最后一个stage设置时长为0，正常
func TestRunnerConfig_Check10(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig1 := NewStage(10*time.Minute, 100, 250)
	stageconfig2 := NewStage(10*time.Minute, 100, 250)
	stageconfig3 := NewStage(0*time.Minute, 10, 250)

	runnerconfig.AppendStages(stageconfig1, stageconfig2, stageconfig3)

	if err := runnerconfig.check(); err != nil {
		t.Error("runnerconfig.Concurrence < 0 ")
	}
}

func TestBaseRunner_WithDeadLine(t *testing.T) {
	duras := []time.Duration{
		3 * time.Second,
		5 * time.Minute,
		10 * time.Hour,
		0 * time.Second,
		0 * time.Minute,
		-1 * time.Second,
		-5 * time.Minute,
		10000 * time.Minute,
		99999 * time.Hour,
	}

	for _, dura := range duras {
		stagerunner := newBaseRunner()
		deadline := time.Now().Add(dura)
		stagerunner.WithDeadLine(deadline)

		if stagerunner.deadline != deadline {
			t.Error("BaseRunner_WithDeadLine wrong")
		}
	}
}

func Test_hatchWorkerCounts(t *testing.T) {
	StageConfigs := StageConfigChanged{30 * time.Second, 500, 100}
	//runnerconfig := NewRunnerConfig()
	//runnerconfig.AppendStages(StageConfigs)
	s1 := []int{100, 100, 100, 100, 100}
	stagecount1 := StageConfigs.hatchWorkerCounts()
	for index, count := range stagecount1 {
		if count != s1[index] {
			t.Error("hatchWorkerChangeCounts wrong")
		}
	}

	StageConfigsChanged2 := StageConfigChanged{30 * time.Second, 1000, 0}
	//runnerconfig2 := NewRunnerConfig()
	//runnerconfig2.AppendStages(StageConfigsChanged2)
	s2 := []int{1000}
	stagecount2 := StageConfigsChanged2.hatchWorkerCounts()
	for index, count := range stagecount2 {
		if count != s2[index] {
			t.Error("hatchWorkerChangeCounts wrong")
		}
	}

	StageConfigsChanged3 := StageConfigChanged{30 * time.Second, 300, 300}
	//runnerconfig3 := NewRunnerConfig()
	//runnerconfig3.AppendStages(StageConfigsChanged3)
	s3 := []int{300}
	stagecount3 := StageConfigsChanged3.hatchWorkerCounts()
	for index, count := range stagecount3 {
		if count != s3[index] {
			t.Error("hatchWorkerChangeCounts wrong")
		}
	}

	StageConfigsChanged4 := StageConfigChanged{30 * time.Second, 500, 300}
	//runnerconfig4 := NewRunnerConfig()
	//runnerconfig4.AppendStages(StageConfigsChanged4)
	s4 := []int{300, 200}
	stagecount4 := StageConfigsChanged4.hatchWorkerCounts()
	for index, count := range stagecount4 {
		if count != s4[index] {
			t.Error("hatchWorkerChangeCounts wrong")
		}
	}

	StageConfigsChanged5 := StageConfigChanged{30 * time.Second, 500, 8000}
	//runnerconfig5 := NewRunnerConfig()
	//runnerconfig5.AppendStages(StageConfigsChanged5)
	s5 := []int{500}
	stagecount5 := StageConfigsChanged5.hatchWorkerCounts()
	for index, count := range stagecount5 {
		if count != s5[index] {
			t.Error("hatchWorkerChangeCounts wrong")
		}
	}

	StageConfigsChanged6 := StageConfigChanged{30 * time.Second, 200, 0}
	s6 := []int{200}
	stagecount6 := StageConfigsChanged6.hatchWorkerCounts()
	for index, count := range stagecount6 {
		if count != s6[index] {
			t.Error("hatchWorkerChangeCounts wrong")
		}
	}
}

func TestBaseRunner_GetStageRunningTime(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig1 := NewStage(10*time.Minute, 100, 250)
	stageconfig2 := NewStage(10*time.Second, 100, 250)
	stageconfig3 := NewStage(1110*time.Hour, 10, 250)

	testdata := []time.Duration{10 * time.Minute, 10 * time.Second, 1110 * time.Hour}

	runnerconfig.AppendStages(stageconfig1, stageconfig2, stageconfig3)
	baseRunner := newBaseRunner()
	baseRunner.WithConfig(runnerconfig)
	runningtime := baseRunner.GetStageRunningTime()

	for i, td := range testdata {
		if runningtime[i] != td {
			t.Error("baseRunner.GetStageRunningTime wrong")
		}
	}
}

func TestBaseRunner_GetStatus(t *testing.T) {
	runnerconfig := NewRunnerConfig()
	stageconfig1 := NewStage(10*time.Minute, 100, 250)
	stageconfig2 := NewStage(10*time.Second, 100, 250)
	stageconfig3 := NewStage(1110*time.Hour, 10, 250)

	//testdata := []time.Duration{10 * time.Minute, 10 * time.Second, 1110 * time.Hour}

	runnerconfig.AppendStages(stageconfig1, stageconfig2, stageconfig3)
	baseRunner := newBaseRunner()

	status1 := baseRunner.GetStatus()
	if status1 != 0 {
		t.Error("baseRunner.GetStatus wrong1")
	}

	baseRunner.Done()
	status2 := baseRunner.GetStatus()
	if status2 != StatusStopped {
		t.Error("baseRunner.GetStatus wrong2")
	}

}

func Test_hatchWorkerCounts2(t *testing.T) {
	stageconfig1 := StageConfigChanged{0 * time.Second, -160, 0}
	stageconfig2 := StageConfigChanged{0 * time.Hour, 270, 16}
	stageconfig3 := StageConfigChanged{0 * time.Hour, -100, 15}
	stageconfig4 := StageConfigChanged{0 * time.Hour, 210, 100}
	stageconfig5 := StageConfigChanged{0 * time.Hour, 210, 0}
	stageconfig6 := StageConfigChanged{0 * time.Hour, 10, 23}
	stageconfig7 := StageConfigChanged{0 * time.Hour, -10, 23}

	ints1 := stageconfig1.hatchWorkerCounts()
	ints2 := stageconfig2.hatchWorkerCounts()
	ints3 := stageconfig3.hatchWorkerCounts()
	ints4 := stageconfig4.hatchWorkerCounts()
	ints5 := stageconfig5.hatchWorkerCounts()
	ints6 := stageconfig6.hatchWorkerCounts()
	ints7 := stageconfig7.hatchWorkerCounts()

	//fmt.Println(ints7)

	var intss = [][]int{
		ints1,
		ints2,
		ints3,
		ints4,
		ints5,
		ints6,
		ints7,
	}

	var testdatas = [][]int{
		{-160},
		{16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 14},
		{-15, -15, -15, -15, -15, -15, -10},
		{100, 100, 10},
		{210},
		{10},
		{-10},
	}

	for ti, testdata := range testdatas {
		fmt.Println(ti, testdata)
		for i, in := range intss[ti] {
			if in != testdata[i] {
				t.Error("stageconfig.hatchWorkerCounts error", intss[ti], testdata)
			}
		}
	}
}

//兼容v1
func TestBaseRunner_startcompatible(t *testing.T) {
	t.Skip("just for debug")

	task := NewTask()
	task.Add(NewHTTPAttacker("multilanguage",
		func() (*http.Request, error) {
			req, _ := http.NewRequest(http.MethodGet, "http://www.baidu.com", nil)
			return req, nil
		}), 10)
	//task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)

	base := newBaseRunner()
	base.WithTask(task)
	base.Config.Concurrence = 100
	base.Config.HatchRate = 10
	base.Config.MinWait = ZeroDuration
	base.Config.MaxWait = ZeroDuration
	//base.Config.Requests = 2000
	base.WithDeadLine(time.Now().Add(2 * time.Minute))
	LocalRunner.baseRunner = base

	LocalRunner.Start()

}

func TestBaseRunner_start(t *testing.T) {
	t.Skip("just for debug ")

	task := NewTask()
	task.Add(NewHTTPAttacker("multilanguage",
		func() (*http.Request, error) {
			req, _ := http.NewRequest(http.MethodGet, "http://www.baidu.com", nil)
			return req, nil
		}), 10)
	//task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)

	stageconfig := NewStage(1*time.Minute, 100, 10)
	stageconfig2 := NewStage(2*time.Minute, 300, 100)
	stageconfig3 := NewStage(ZeroDuration, 150, 10)
	runnerconfig := NewRunnerConfig()
	runnerconfig.AppendStages(stageconfig).AppendStages(stageconfig2, stageconfig3)
	//runnerconfig.Requests = 8000
	// .AppendStages(stageconfig3)

	base := newBaseRunner()
	base.WithTask(task)
	//base.WithDeadLine(time.Now().Add(3 *time.Minute))
	base.WithConfig(runnerconfig)
	LocalRunner.baseRunner = base

	Logger.Info("baserunnr: ", zap.Any("info", LocalRunner.baseRunner))
	Logger.Info("baserunnr: ", zap.Any("info", LocalRunner.baseRunner.deadline))

	LocalRunner.Start()
}

func TestBaseRunner_start2(t *testing.T) {
	t.Skip("just for debug ")

	task := NewTask()
	task.Add(NewHTTPAttacker("multilanguage",
		func() (*http.Request, error) {
			req, _ := http.NewRequest(http.MethodGet, "http://www.baidu.com", nil)
			return req, nil
		}), 10)
	//task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)

	stageconfig := NewStage(1*time.Minute, 100, 10)
	stageconfig2 := NewStage(1*time.Minute, 200, 0)
	//stageconfig3 := NewStage(2 * time.Minute, 150, 10)
	runnerconfig := NewRunnerConfig()
	runnerconfig.AppendStages(stageconfig).AppendStages(stageconfig2)
	// .AppendStages(stageconfig3)

	base := newBaseRunner()
	base.WithTask(task)
	base.WithConfig(runnerconfig)
	LocalRunner.baseRunner = base

	LocalRunner.Start()

}

func TestBaseRunner_start3(t *testing.T) {
	t.Skip("just for debug ")

	task := NewTask()
	task.Add(NewHTTPAttacker("multilanguage",
		func() (*http.Request, error) {
			req, _ := http.NewRequest(http.MethodGet, "http://shouqianba-multilanguage.test.shouqianba.com/app/languages?appkey=ws_1540346060991", nil)
			return req, nil
		}), 10)
	//task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)

	//stageconfig := NewStage(1 * time.Minute, 100, 10)
	//stageconfig2 := NewStage(0 *time.Minute, 200, 0)
	//stageconfig3 := NewStage(2 * time.Minute, 150, 10)
	runnerconfig := NewRunnerConfig()
	runnerconfig.Duration = 0 * time.Minute
	runnerconfig.Concurrence = 100
	runnerconfig.HatchRate = 10
	//runnerconfig.AppendStages(stageconfig).AppendStages(stageconfig2)
	// .AppendStages(stageconfig3)

	base := newBaseRunner()
	base.WithTask(task)
	base.WithDeadLine(time.Now().Add(4 * time.Minute))
	base.WithConfig(runnerconfig)
	LocalRunner.baseRunner = base

	LocalRunner.Start()

}

//定义stage及v1runner，报错
func TestBaseRunner_start4(t *testing.T) {
	t.Skip("just for debug ")

	task := NewTask()
	task.Add(NewHTTPAttacker("multilanguage",
		func() (*http.Request, error) {
			req, _ := http.NewRequest(http.MethodGet, "http://shouqianba-multilanguage.test.shouqianba.com/app/languages?appkey=ws_1540346060991", nil)
			return req, nil
		}), 10)
	//task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)

	stageconfig := NewStage(1*time.Minute, 200, 10)
	stageconfig2 := NewStage(1*time.Minute, 100, 0)
	//stageconfig3 := NewStage(2 * time.Minute, 150, 10)
	runnerconfig := NewRunnerConfig()
	runnerconfig.Duration = 0 * time.Minute
	runnerconfig.Concurrence = 100
	runnerconfig.HatchRate = 10
	runnerconfig.AppendStages(stageconfig).AppendStages(stageconfig2)
	// .AppendStages(stageconfig3)

	base := newBaseRunner()
	base.WithTask(task)
	base.WithDeadLine(time.Now().Add(4 * time.Minute))
	base.WithConfig(runnerconfig)
	LocalRunner.baseRunner = base

	LocalRunner.Start()

}

//定义总请求数
func TestBaseRunner_start5(t *testing.T) {
	t.Skip("just for debug ")

	task := NewTask()
	task.Add(NewHTTPAttacker("multilanguage",
		func() (*http.Request, error) {
			req, _ := http.NewRequest(http.MethodGet, "http://www.baidu.com", nil)
			return req, nil
		}), 10)
	//task.Add(newAttacker("b"), 20)
	task.Add(newAttacker("c"), 3)

	stageconfig := NewStage(1*time.Minute, 200, 10)
	stageconfig2 := NewStage(1*time.Minute, 100, 0)
	//stageconfig3 := NewStage(2 * time.Minute, 150, 10)
	runnerconfig := NewRunnerConfig()
	runnerconfig.Requests = 1000
	runnerconfig.AppendStages(stageconfig).AppendStages(stageconfig2)
	// .AppendStages(stageconfig3)

	base := newBaseRunner()
	base.WithTask(task)
	base.WithDeadLine(time.Now().Add(1 * time.Minute))
	base.WithConfig(runnerconfig)
	LocalRunner.baseRunner = base

	LocalRunner.Start()

}

func TestBaseRunner_UpdateDeadline(t *testing.T) {
	br := newBaseRunner()
	rc := NewRunnerConfig()
	sc1 := NewStage(10*time.Minute, 100, 30)
	sc2 := NewStage(1*time.Minute, 100, 30)
	sc3 := NewStage(2*time.Minute, 100, 30)
	rc.AppendStages(sc1, sc2, sc3)
	rc.updateStageConfig()
	br.WithConfig(rc)
	br.updateDeadline()

	if !(br.deadline.After(time.Now().Add(12*time.Minute)) && br.deadline.Before(time.Now().Add(14*time.Minute))) {
		t.Error("UpdateDeadline error")
	}
}

func TestBaseRunner_UpdateDeadline1(t *testing.T) {
	br := newBaseRunner()
	rc := NewRunnerConfig()
	sc1 := NewStage(1*time.Hour, 100, 30)
	sc2 := NewStage(1*time.Minute, 100, 30)
	sc3 := NewStage(2*time.Minute, 100, 30)
	rc.AppendStages(sc1, sc2, sc3)
	rc.updateStageConfig()
	br.WithConfig(rc)
	br.updateDeadline()

	if !(br.deadline.After(time.Now().Add(62*time.Minute)) && br.deadline.Before(time.Now().Add(64*time.Minute))) {
		t.Error("UpdateDeadline error")
	}
}

func TestBaseRunner_UpdateDeadline2(t *testing.T) {
	br := newBaseRunner()
	rc := NewRunnerConfig()
	sc1 := NewStage(10*time.Minute, 100, 30)
	sc2 := NewStage(ZeroDuration, 100, 30)
	sc3 := NewStage(2*time.Minute, 100, 30)
	rc.AppendStages(sc1, sc2, sc3)
	rc.updateStageConfig()
	br.WithConfig(rc)
	br.updateDeadline()

	if !br.deadline.IsZero() {
		t.Error("UpdateDeadline error")
	}
}

func TestBaseRunner_UpdateBaseRunner(t *testing.T) {

	br := newBaseRunner()
	rc := NewRunnerConfig()
	sc1 := NewStage(10*time.Minute, 100, 30)
	//sc2 := NewStage(ZeroDuration, 100, 30)
	sc3 := NewStage(2*time.Minute, 100, 30)
	rc.AppendStages(sc1, sc3)
	br.WithConfig(rc)

	br.activeBaseRunner()

	if !br.deadline.After(time.Now().Add(11*time.Minute)) && br.deadline.Before(time.Now().Add(13*time.Minute)) {
		t.Error("updateBaseRunner error")
	}
}

func TestBaseRunner_UpdateBaseRunner2(t *testing.T) {

	br := newBaseRunner()
	rc := NewRunnerConfig()
	sc1 := NewStage(10*time.Minute, 100, 30)
	//sc2 := NewStage(ZeroDuration, 100, 30)
	sc3 := NewStage(0*time.Minute, 100, 30)
	rc.AppendStages(sc1, sc3)
	br.WithConfig(rc)

	br.activeBaseRunner()

	if !br.deadline.Equal(time.Time{}) {
		t.Error("updateBaseRunner error")
	}
}

func BenchmarkBaseRunner_AddCancelFunc(b *testing.B) {
	_, cancel := context.WithCancel(context.Background())
	br := newBaseRunner()
	for i := 0; i < b.N; i++ {
		br.AddCancelFunc(&cancel)
	}
}
