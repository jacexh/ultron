package ultron

import "time"

func timeDurationToRoudedMillisecond(t time.Duration) roundedMillisecond {
	ms := int64(t.Seconds()*1000 + 0.5)
	var rm roundedMillisecond
	if ms < 100 {
		rm = roundedMillisecond(ms)
	} else if ms < 1000 {
		rm = roundedMillisecond(ms + 5 - (ms+5)%10)
	} else {
		rm = roundedMillisecond((ms + 50 - (ms+50)%100))
	}
	return rm
}

func roundedMillisecondToDuration(r roundedMillisecond) time.Duration {
	ret := time.Duration(r * 1000 * 1000)
	return ret
}

func timeDurationToMillsecond(t time.Duration) int64 {
	return int64(t) / int64(time.Millisecond)
}
