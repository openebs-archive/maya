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
	"os/exec"
	"strings"
	"sync"
)

// RunCommand executes the command supplied and return the error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommand(cmd string) error {
	logger.PrintfDebugMessage("Executing command: %q\n", cmd)

	// splitting head => g++ parts => rest of the command
	// python equivalent: parts = [x.strip() for x in cmd.split() if x.strip()]
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	err := exec.Command(head, parts...).Run()
	return err
}

// RunCommandWithSudo executes the command supplied and return the error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandWithSudo(cmd string) error {
	return RunCommand("sudo " + cmd)
}

// RunCommandSync executes the command supplied and return the error
// It also takes one sync.WaitGroup object as an argument which it notifies once command is executed
func RunCommandSync(cmd string, wg *sync.WaitGroup) (string, error) {
	err := RunCommand(cmd)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return "", err
}

// runCommandArrayWithGivenStdin is the internal function which does the core job of
// all runCommand*WithGivenStdin* functions
func runCommandArrayWithGivenStdin(head string, parts []string, stdin string) error {
	command := exec.Command(head, parts...)

	command.Stdin = bytes.NewBuffer([]byte(stdin))

	err := command.Run()
	return err
}

// RunCommandArrayWithGivenStdin runs the command supplied
// then feed the supplied stdin to commands stdin then return the error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandArrayWithGivenStdin(cmd []string, stdin string) error {
	return runCommandArrayWithGivenStdin(cmd[0], cmd[1:], stdin)
}

// RunCommandWithGivenStdin runs the command supplied
// then feed the supplied stdin to commands stdin then return error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandWithGivenStdin(cmd, stdin string) error {
	return RunCommandArrayWithGivenStdin(strings.Fields(cmd), stdin)
}

// RunCommandArrayWithGivenStdinWithSudo runs the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the error
// don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandArrayWithGivenStdinWithSudo(cmd []string, stdin string) error {
	return runCommandArrayWithGivenStdin("sudo", cmd, stdin)
}

// RunCommandWithGivenStdinWithSudo runs the command supplied with `sudo`
// then feed the supplied stdin to commands stdin then return the error
// Since this function splits the commands on whitespaces, avoid those commands
// whose argument has space or if the command itself have the space
// Also don't use quotes in command or argument because that quote will be considered
// part of the command
func RunCommandWithGivenStdinWithSudo(cmd, stdin string) error {
	return RunCommandArrayWithGivenStdinWithSudo(strings.Fields(cmd), stdin)
}
