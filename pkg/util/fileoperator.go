package util

import (
	"io/ioutil"
	"os"

	"github.com/golang/glog"
)

//FileOperator operates on files
type FileOperator interface {
	Write(filename string, data []byte, perm os.FileMode) error
}

//RealFileOperator is used for writing the actual files without mocking
type RealFileOperator struct{}

// the real file operator for the actual program,
func (r RealFileOperator) Write(filename string, data []byte, perm os.FileMode) error {
	err := ioutil.WriteFile(filename, data, perm)
	if err != nil {
		glog.Errorf("Failed to write file: " + filename)
	}
	return err
}

//TestFileOperator is used as a dummy FileOperator
type TestFileOperator struct{}

//Write is to mock write operation for FileOperator interface
func (r TestFileOperator) Write(filename string, data []byte, perm os.FileMode) error {
	return nil
}
