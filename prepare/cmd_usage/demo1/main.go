package main

import (
	"fmt"
	"os/exec"
)

//exec包的最简单示例
func main() {
	cmd := exec.Command("F:\\cygwin64\\bin\\bash.exe", "-c", "echo hello")
	err := cmd.Run()
	fmt.Println(err)
}
