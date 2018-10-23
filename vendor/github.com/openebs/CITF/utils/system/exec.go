/*
Copyright 2018 The OpenEBS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package system

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/openebs/CITF/utils/log"
)

var logger log.Logger

// ExecCommand executes the command supplied and return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommand(cmd string) (string, error) {
	logger.PrintfDebugMessage("Executing command: %q\n", cmd)

	// splitting head => g++ parts => rest of the command
	// python equivalent: parts = [x.strip() for x in cmd.split() if x.strip()]
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	out, err := exec.Command(head, parts...).Output()
	return string(out), err
}

// ExecCommandSync executes the command supplied and return the output as well as error
// It also takes one sync.WaitGroup object as an argument which it notifies once command is executed
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandSync(cmd string, wg *sync.WaitGroup) (string, error) {
	out, err := ExecCommand(cmd)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return string(out), err
}

// ExecCommandWithSudo executes the command supplied with `sudo` and return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithSudo(cmd string) (string, error) {
	return ExecCommand("sudo " + cmd)
}

// ExecCommandWithSudoSync executes the command supplied with `sudo` and return the output as well as error
// It also takes one sync.WaitGroup object as an argument which it notifies once command is executed
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithSudoSync(cmd string, wg *sync.WaitGroup) (string, error) {
	out, err := ExecCommandWithSudo(cmd)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return string(out), err
}

// execCommandArrayWithGivenStdin is the internal function which does the core job of
// all ExecCommand*WithGivenStdin* functions
func execCommandArrayWithGivenStdin(head string, parts []string, stdin string) (string, error) {
	command := exec.Command(head, parts...)

	command.Stdin = bytes.NewBuffer([]byte(stdin))

	out, err := command.Output()
	return string(out), err
}

// ExecCommandArrayWithGivenStdin executes the command supplied
// then feed the supplied stdin to commands stdin then return the output as well as error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandArrayWithGivenStdin(cmd []string, stdin string) (string, error) {
	return execCommandArrayWithGivenStdin(cmd[0], cmd[1:], stdin)
}

// ExecCommandWithGivenStdin executes the command supplied
// then feed the supplied stdin to commands stdin then return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithGivenStdin(cmd, stdin string) (string, error) {
	return ExecCommandArrayWithGivenStdin(strings.Fields(cmd), stdin)
}

// ExecCommandArrayWithGivenStdinWithSudo executes the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the output as well as error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandArrayWithGivenStdinWithSudo(cmd []string, stdin string) (string, error) {
	return execCommandArrayWithGivenStdin("sudo", cmd, stdin)
}

// ExecCommandWithGivenStdinWithSudo executes the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the output as well as error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecCommandWithGivenStdinWithSudo(cmd, stdin string) (string, error) {
	return ExecCommandArrayWithGivenStdinWithSudo(strings.Fields(cmd), stdin)
}

// ExecPipeTwoCommandsArray takes two commands in its parameter. It runs first command
// and feed its output to second command as input
func ExecPipeTwoCommandsArray(cmd1, cmd2 []string) (string, error) {
	logger.PrintfDebugMessage("Executing command: %q\n", strings.Join(cmd1, " ")+" | "+strings.Join(cmd2, " "))

	c1 := exec.Command(cmd1[0], cmd1[1:]...)
	c2 := exec.Command(cmd2[0], cmd2[1:]...)

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	// var err error
	// c2.Stdin, err = c1.StdoutPipe()
	// if err != nil {
	// 	return "", fmt.Errorf("Error getting stdout pipe of command: %q. Error: %+v", cmd1, err)
	// }

	var b2 bytes.Buffer
	c2.Stdout = &b2

	err := c1.Start()
	if err != nil {
		return "", fmt.Errorf("error starting command: %q. Error: %+v", cmd1, err)
	}
	err = c2.Start()
	if err != nil {
		return "", fmt.Errorf("error starting command: %q. Error: %+v", cmd2, err)
	}
	err = c1.Wait()
	if err != nil {
		return "", fmt.Errorf("error while waiting for command: %q to exit. Error: %+v", cmd1, err)
	}
	err = w.Close()
	if err != nil {
		return "", fmt.Errorf("error while closing the pipe writer. Error: %+v", err)
	}
	err = c2.Wait()
	if err != nil {
		return "", fmt.Errorf("error while waiting for command: %q to exit. Error: %+v", cmd2, err)
	}

	return b2.String(), nil
}

// ExecPipeTwoCommands takes two commands in its parameter. It runs first command
// and feed its output to second command as input
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func ExecPipeTwoCommands(cmd1, cmd2 string) (string, error) {
	// splitting head => g++ parts => rest of the command
	// python equivalent: parts = [x.strip() for x in cmd.split() if x.strip()]
	parts1 := strings.Fields(cmd1)
	parts2 := strings.Fields(cmd2)

	return ExecPipeTwoCommandsArray(parts1, parts2)
}
