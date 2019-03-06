// Copyright © 2018-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Updatefile(filename, updateStorageVal string, index int, perm os.FileMode) error
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
func (r RealFileOperator) Updatefile(filename, updateStorageVal string, index int, perm os.FileMode) error {
	buffer, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to read a %s file", filename)
	}
	lines := strings.Split(string(buffer), "\n")
	lines[index] = updateStorageVal
	newbuffer := strings.Join(lines, "\n")
	glog.Infof("Values in lines %s\n", lines)
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
