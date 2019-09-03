/*
Copyright 2019 The OpenEBS Authors.

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

package pstatus

import (
	"bytes"
	"strings"
	"unicode"

	vdump "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
)

// SetPool method set the Pool field of PoolDump object.
func (p *PoolDump) SetPool(Pool string) {
	p.Pool = Pool
}

// SetCommand method set the Command field of PoolDump object.
func (p *PoolDump) SetCommand(Command string) {
	p.Command = Command
}

// GetPool method get the Pool field of PoolDump object.
func (p *PoolDump) GetPool() string {
	return p.Pool
}

// GetCommand method get the Command field of PoolDump object.
func (p *PoolDump) GetCommand() string {
	return p.Command
}

func stripDiskPath(vdevlist []vdump.Vdev) {
	for i, v := range vdevlist {
		if v.IsWholeDisk == 1 {
			vdevlist[i].Path = getDiskStripPath(v.Path)
		}
		stripDiskPath(v.Children)
	}
}

// getLastIndex return the index which doesn't satisfy
// characteristics of given function fn otherwise
// it will return the length of the string
func getLastIndex(p []byte, fn func(r rune) bool) int {
	for idx := range p {
		if !fn(rune(p[idx])) {
			return idx
		}
	}
	return len(p)
}

/* getDiskStripPath Remove partition suffix from a vdev path.
 * Partition suffixes may take three forms:
 * 1. "-partX", "pX", or "X", where X is a string of digits,
 *    like /dev/disk/by-id/scsi-0Google_PersistentDisk_persistent-disk-0-part1
 *
 * 2. when the suffix is preceded by a digit, i.e. "md0p0",
 *    like /dev/md0p0
 *
 * 3. when preceded by a string matching the regular expression
 *    "^([hsv]|xv)d[a-z]+", i.e. a scsi, ide, virtio or xen disk,
 *    like, /dev/xvdlps3, /dev/hdvdas2, /dev/sda1
 */
func getDiskStripPath(path string) string {
	var part, d []byte

	if len(path) == 0 {
		return path
	}
	npathBytes := []byte(path)
	lastIndex := bytes.LastIndexByte(npathBytes, byte('/'))
	if lastIndex == -1 {
		return path
	}
	pathBytes := npathBytes[lastIndex+1:]

	if idx := strings.Index(string(pathBytes), "-part"); idx != -1 && idx != 0 {
		part = pathBytes[idx:]
		d = part[5:]
	} else if idx = bytes.LastIndexByte(pathBytes, byte('p')); idx != -1 && pathBytes[idx] > pathBytes[1] && unicode.IsNumber(rune(pathBytes[idx-1])) {
		part = pathBytes[idx:]
		d = part[1:]
	} else if bytes.ContainsAny([]byte(string(pathBytes[0])), "hsv") && pathBytes[1] == 'd' {
		d = pathBytes[2:]
		d = d[getLastIndex(d, unicode.IsLetter):]
		part = d
	} else if bytes.Equal(pathBytes[:3], []byte("xvd")) {
		d = pathBytes[3:]
		d = d[getLastIndex(d, unicode.IsLetter):]
		part = d
	}

	if len(part) != 0 && len(d) != 0 {
		d = d[getLastIndex(d, unicode.IsNumber):]
		if len(d) == 0 {
			for i := range part {
				part[i] = byte('\n')
			}
		}
	}
	return strings.Split(string(npathBytes), "\n")[0]
}
