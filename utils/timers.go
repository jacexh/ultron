package utils

import (
	"time"
)

type Tis struct {
	d []time.Duration
	C chan int
}

//func timers(t *Tis) {
//	var index = 0
//	t.C <- index
//	for num, d := range t.d {
//		if num == len(t.d)-1 && d == time.Duration(0){
//			//do nothing
//		} else {
//			index ++
//			<- time.NewTimer(d).C
//			t.C <- index
//		}
//	}
//	for {
//		time.Sleep(1 * time.Hour)
//	}
//}

func timers(t *Tis) {
	//var timer = time.NewTimer(time.Second)
	var index = 0
	t.C <- index
	for _, d := range t.d {

		//timer.Reset(d)
		//<-timer.C
		//index ++
		//t.C <- index
		if d != time.Duration(0) {
			time.Sleep(d)
			index ++
			t.C <- index
		}

	}
}


//TODO 使用reset()
// 类似time包里的newtimer。只不过这个支持多个时间节点。
func NewTimers(d []time.Duration) *Tis {
	c := make(chan int)
	t := &Tis{
		d: d,
		C: c,
	}
	go timers(t)
	return t
}




