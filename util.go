package ultron

import "time"

func timeDurationToRoudedMillisecond(t time.Duration) RoundedMillisecond {
	ms := int64(t.Seconds()*1000 + 0.5)
	var rm RoundedMillisecond
	if ms < 100 {
		rm = RoundedMillisecond(ms)
	} else if ms < 1000 {
		rm = RoundedMillisecond(ms + 5 - (ms+5)%10)
	} else {
		rm = RoundedMillisecond((ms + 50 - (ms+50)%100))
	}
	return rm
}

func roundedMillisecondToDuration(r RoundedMillisecond) time.Duration {
	return time.Duration(r * 1000 * 1000)
}

func timeDurationToMillsecond(t time.Duration) int64 {
	return int64(t) / int64(time.Millisecond)
}
