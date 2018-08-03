package util

import (
	"io/ioutil"
	"os/exec"

	"github.com/golang/glog"
)

// Runner interface implements various methods of running binaries which can be
// modified for unit testing.
type Runner interface {
	RunCombinedOutput(string, ...string) ([]byte, error)
	RunStdoutPipe(string, ...string) ([]byte, error)
}

// RealRunner is the real runner for the program that actually execs the command.
type RealRunner struct{}

// RunCombinedOutput runs the command and returns its combined standard output
// and standard error.
func (r RealRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	//execute pool creation command.
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

// RunStdoutPipe returns a pipe that will be connected to the command's standard output
// when the command starts.
func (r RealRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		glog.Errorf(err.Error())
		return []byte{}, err
	}
	if err := cmd.Start(); err != nil {
		glog.Errorf(err.Error())
		return []byte{}, err
	}
	data, _ := ioutil.ReadAll(stdout)
	if err := cmd.Wait(); err != nil {
		glog.Errorf(err.Error())
		return []byte{}, err
	}
	return data, nil
}

//TestRunner is used as a dummy Runner
type TestRunner struct{}

// RunCombinedOutput is to mock Real runner exec.
func (r TestRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}

// RunStdoutPipe is to mock real runner exec with stdoutpipe.
func (r TestRunner) RunStdoutPipe(command string, args ...string) ([]byte, error) {
	return []byte("success"), nil
}
