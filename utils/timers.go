package utils

import (
	//"Richard1ybb/ultron"
	"time"
)

type tis struct {
	d []time.Duration
	C chan int
}

func timers(t *tis) {
	var index = 0
	t.C <- index
	for num, d := range t.d {
		if num == len(t.d)-1 && d == time.Duration(0){
			//do nothing
		} else {
			index ++
			<- time.NewTimer(d).C
			t.C <- index
		}
	}
	for {
		time.Sleep(1 * time.Hour)
	}
}

// 类似time包里的newtimer。只不过这个支持多个时间节点。
func NewTimers(d []time.Duration) *tis {
	c := make(chan int)
	t := &tis{
		d: d,
		C: c,
	}
	go timers(t)
	return t
}




