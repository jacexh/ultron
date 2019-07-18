package ultron

import (
	"fmt"
	"time"
)

type Tis struct {
	d []time.Duration
	C chan int
}

func timers(t *Tis) {
	//var timer = time.NewTimer(time.Second)
	var index = 0
	t.C <- index
	for num, d := range t.d {

		//timer.Reset(d)
		//<-timer.C
		//index ++
		//t.C <- index
		if d != time.Duration(0) {
			if num >= len(t.d)-1 {
				time.Sleep(d)
				t.C <- -1
			} else {
				time.Sleep(d)
				index++
				t.C <- index
			}
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

func abs(i int) int {
	if i < 0 {
		return -i
	} else {
		return i
	}
}

func ShowLogo() {
	fmt.Println(`
      ___           ___       ___           ___           ___           ___     
     /\__\         /\__\     /\  \         /\  \         /\  \         /\__\    
    /:/  /        /:/  /     \:\  \       /::\  \       /::\  \       /::|  |   
   /:/  /        /:/  /       \:\  \     /:/\:\  \     /:/\:\  \     /:|:|  |   
  /:/  /  ___   /:/  /        /::\  \   /::\~\:\  \   /:/  \:\  \   /:/|:|  |__ 
 /:/__/  /\__\ /:/__/        /:/\:\__\ /:/\:\ \:\__\ /:/__/ \:\__\ /:/ |:| /\__\
 \:\  \ /:/  / \:\  \       /:/  \/__/ \/_|::\/:/  / \:\  \ /:/  / \/__|:|/:/  /
  \:\  /:/  /   \:\  \     /:/  /         |:|::/  /   \:\  /:/  /      |:/:/  / 
   \:\/:/  /     \:\  \    \/__/          |:|\/__/     \:\/:/  /       |::/  /  
    \::/  /       \:\__\                  |:|  |        \::/  /        /:/  /   
     \/__/         \/__/                   \|__|         \/__/         \/__/    
`)
}
