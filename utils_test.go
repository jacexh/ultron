package ultron

import "testing"

func ExampleShowLogo() {
	ShowLogo()
	//Output:
	//      _  _
	//  /\ /\ | || |_  _ __   ___   _ __
	/// / \ \| || __|| '__| / _ \ | '_ \
	//\ \_/ /| || |_ | |   | (_) || | | |
	//  \___/ |_| \__||_|    \___/ |_| |_|

}

func TestAbs(t *testing.T) {
	type testData struct{
		in   int
		want int
	}

	var tds = []testData{
		{0, 0},
		{1, 1},
		{-1,1},
		{-999, 999},
		{999, 999},
	}

	for _, td := range tds {
		if got := Abs(td.in); td.want != got {
			t.Errorf("Abs(): got: %s, want: %s", got, td.want)
		}
	}
}
