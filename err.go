package ultron

import "time"

func newAttackerError(a string, err error) *AttackerError {
	return &AttackerError{Name: a, CausedBy: err.Error()}
}

func (ae *AttackerError) Error() string {
	return ae.CausedBy
}

func newResult(n string, d time.Duration, err error) *Result {
	if err == nil {
		return &Result{Name: n, Duration: int64(d)}
	}
	return &Result{Name: n, Duration: int64(d), Error: newAttackerError(n, err)}
}
