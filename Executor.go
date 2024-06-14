package shared

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ExecuteShellCommand(shellCommand string) error {
	return exec.Command("/bin/sh", "-c", shellCommand).Run()
}

func Execute(commandStr string) {
	commandParts := strings.Split(commandStr, " ")
	command := exec.Command(commandParts[0], commandParts[1:]...)
	err := command.Run()
	if err != nil {
		fmt.Printf("Error executing docker command: %v\n", err)
		os.Exit(1)
	}
}
