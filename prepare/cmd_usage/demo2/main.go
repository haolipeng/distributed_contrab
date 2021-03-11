package main

import (
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("F:\\cygwin64\\bin\\bash.exe", "-c", "echo hello")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("command exec failed:%s\n", err.Error())
		return
	}
	fmt.Println("output:", string(output))
}
