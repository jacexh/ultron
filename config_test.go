package ultron

import (
	"fmt"
	"time"
)

func ExampleStageConfig_split() {
		sc := NewStageConfig(10 *time.Minute, 100, 30)
	scs := sc.split(4)
	for _, scc := range scs {
		fmt.Println(scc)
	}
	//Output:&{10m0s 25 7}
	//&{10m0s 25 7}
	//&{10m0s 25 7}
	//&{10m0s 25 9}
}

func ExampleStageConfig_split2() {
	sc := NewStageConfig(10 *time.Minute, 100, 30)
	scs := sc.split(9)
	for _, scc := range scs {
		fmt.Println(scc)
	}
	//Output:&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 11 3}
	//&{10m0s 12 6}
}

func ExampleStageConfig_split3() {
	sc := NewStageConfig(10 *time.Minute, 100, 30)
	scs := sc.split(1)
	for _, scc := range scs {
		fmt.Println(scc)
	}
	//Output:&{10m0s 100 30}

}

func ExampleRunnerConfig_split4() {
	sc1 := NewStageConfig(10 *time.Minute, 100, 30)
	sc2 := NewStageConfig(10 *time.Minute, 70, 18)
	rc := NewRunnerConfig()
	rc.AppendStage(sc1, sc2)

	rcs := rc.split(4)

	for _, rc := range rcs {
		//fmt.Println(rc)
		for _, r := range rc.Stages {
			fmt.Println(r)
		}
	}
	//Output:
	//&{10m0s 25 7}
	//&{10m0s 17 4}
	//&{10m0s 25 7}
	//&{10m0s 17 4}
	//&{10m0s 25 7}
	//&{10m0s 17 4}
	//&{10m0s 25 9}
	//&{10m0s 19 6}

}

func ExampleRunnerConfig_split5() {
	sc1 := NewStageConfig(10*time.Minute, 100, 0)
	sc2 := NewStageConfig(2*time.Minute, 70, 0)
	rc := NewRunnerConfig()
	rc.AppendStage(sc1, sc2)

	rcs := rc.split(3)

	//fmt.Println(rcs)
	for _, rc := range rcs {
		//fmt.Println(rc)
		for _, r := range rc.Stages {
			fmt.Println(r)
		}
	}
	//Output:
	//&{10m0s 33 0}
	//&{2m0s 23 0}
	//&{10m0s 33 0}
	//&{2m0s 23 0}
	//&{10m0s 34 0}
	//&{2m0s 24 0}

}


func ExampleRunnerConfig_split6() {
	sc1 := NewStageConfig(10*time.Minute, 100, 0)
	//sc2 := NewStageConfig(2*time.Minute, 70, 0)
	rc := NewRunnerConfig()
	rc.AppendStage(sc1)
	rc.Requests = 1000000

	rcs := rc.split(3)

	//fmt.Println(rcs)
	for _, rc := range rcs {
		fmt.Println(rc.Requests)
		for _, r := range rc.Stages {
			fmt.Println(r)
		}
	}

	//Output:
	//333333
	//&{10m0s 33 0}
	//333333
	//&{10m0s 33 0}
	//333334
	//&{10m0s 34 0}
}


func ExampleRunnerConfig_split7() {
	sc1 := NewStageConfig(10*time.Minute, 100, 220)
	//sc2 := NewStageConfig(2*time.Minute, 70, 0)
	rc := NewRunnerConfig()
	rc.AppendStage(sc1)
	rc.Requests = 1000000

	rcs := rc.split(3)

	//fmt.Println(rcs)
	for _, rc := range rcs {
		fmt.Println(rc.Requests)
		for _, r := range rc.Stages {
			fmt.Println(r)
		}
	}

	//Output:
	//333333
	//&{10m0s 33 73}
	//333333
	//&{10m0s 33 73}
	//333334
	//&{10m0s 34 74}
}




