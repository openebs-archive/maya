package util

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//FileOperator operates on files
type FileOperator interface {
	Write(filename string, data []byte, perm os.FileMode) error
	Updatefile(filename, updateVal string, index int, perm os.FileMode) error
	GetLineDetails(filename, searchString string) (int, string, error)
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

// GetLineDetails return the line number and line content of matched string in file
func (r RealFileOperator) GetLineDetails(filename, searchString string) (int, string, error) {
	var line string
	var i int
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return -1, "", errors.Wrapf(err, "failed to read a %s file", filename)
	}
	lines := strings.Split(string(buffer), "\n")
	for i, line = range lines {
		if strings.Contains(line, searchString) {
			return i, line, nil
		}
	}
	return -1, "", nil
}

// Updatefile updates the line number with the given string
func (r RealFileOperator) Updatefile(filename, updateVal string, index int, perm os.FileMode) error {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to read a %s file", filename)
	}
	lines := strings.Split(string(buffer), "\n")
	lines[index] = updateVal
	newbuffer := strings.Join(lines, "\n")
	glog.V(4).Infof("content in a file %s\n", lines)
	err = r.Write(filename, []byte(newbuffer), perm)
	return err
}

//TestFileOperator is used as a dummy FileOperator
type TestFileOperator struct{}

//Write is to mock write operation for FileOperator interface
func (r TestFileOperator) Write(filename string, data []byte, perm os.FileMode) error {
	return nil
}

//Updatefile is to mock Updatefile operation for FileOperator interface
func (r TestFileOperator) Updatefile(filename, updateStorageVal string, index int, perm os.FileMode) error {
	return nil
}

//GetLineDetails is to mock operation for FileOperator interface
func (r TestFileOperator) GetLineDetails(filename, searchString string) (int, string, error) {
	return -1, "", nil
}
