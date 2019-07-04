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

package padd

import "fmt"

// SetNewVdev method set the NewVdev field of PoolDiskReplace object.
func (p *PoolDiskReplace) SetNewVdev(NewVdev string) {
	p.NewVdev = NewVdev
}

// SetOldVdev method set the OldVdev field of PoolDiskReplace object.
func (p *PoolDiskReplace) SetOldVdev(OldVdev string) {
	p.OldVdev = OldVdev
}

// SetProperty method set the Property field of PoolDiskReplace object.
func (p *PoolDiskReplace) SetProperty(key, value string) {
	p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetPool method set the Pool field of PoolDiskReplace object.
func (p *PoolDiskReplace) SetPool(Pool string) {
	p.Pool = Pool
}

// SetCommand method set the Command field of PoolDiskReplace object.
func (p *PoolDiskReplace) SetCommand(Command string) {
	p.Command = Command
}

// GetProperty method get the Property field of PoolDiskReplace object.
func (p *PoolDiskReplace) GetProperty() []string {
	return p.Property
}

// GetPool method get the Pool field of PoolDiskReplace object.
func (p *PoolDiskReplace) GetPool() string {
	return p.Pool
}

// GetCommand method get the Command field of PoolDiskReplace object.
func (p *PoolDiskReplace) GetCommand() string {
	return p.Command
}
