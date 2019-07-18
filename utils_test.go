package ultron

import "testing"

func TestAbs(t *testing.T) {
	type testData struct {
		in   int
		want int
	}

	var tds = []testData{
		{0, 0},
		{1, 1},
		{-1, 1},
		{-999, 999},
		{999, 999},
	}

	for _, td := range tds {
		if got := abs(td.in); td.want != got {
			t.Errorf("abs(): got: %d, want: %d", got, td.want)
		}
	}
}
