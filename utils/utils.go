package utils

import (
	"fmt"
)

func Abs(int2 int) int {
	if int2 < 0 {
		return -int2
	} else {
		return int2
	}
}

func ShowLogo() {
	fmt.Println("        _  _                       ")
	fmt.Println(" /\\ /\\ | || |_  _ __   ___   _ __  ")
	fmt.Println("/ / \\ \\| || __|| '__| / _ \\ | '_ \\ ")
	fmt.Println("\\ \\_/ /| || |_ | |   | (_) || | | |")
	fmt.Println(" \\___/ |_| \\__||_|    \\___/ |_| |_|")
	fmt.Println("                                   ")
}

