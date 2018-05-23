package util

import (
	"io/ioutil"
	"os/exec"

	"github.com/golang/glog"
)

type Runner interface {
	RunCombinedOutput(string, ...string) ([]byte, error)
	RunStdoutPipe(string, ...string) ([]byte, error)
}

type RealRunner struct{}

// the real runner for the actual program, actually execs the command
func (r RealRunner) RunCombinedOutput(command string, args ...string) ([]byte, error) {
	//execute pool creation command.
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

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
