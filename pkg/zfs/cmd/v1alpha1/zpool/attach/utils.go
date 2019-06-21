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

package pattach

import "fmt"

// SetProperty method set the Property field of PoolAttach object.
func (p *PoolAttach) SetProperty(key, value string) {
	p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetForcefully method set the Forcefully field of PoolAttach object.
func (p *PoolAttach) SetForcefully(Forcefully bool) {
	p.Forcefully = Forcefully
}

// SetDevice method set the Device field of PoolAttach object.
func (p *PoolAttach) SetDevice(Device string) {
	p.Device = Device
}

// SetNewDevice method set the NewDevice field of PoolAttach object.
func (p *PoolAttach) SetNewDevice(NewDevice string) {
	p.NewDevice = NewDevice
}

// SetPool method set the Pool field of PoolAttach object.
func (p *PoolAttach) SetPool(Pool string) {
	p.Pool = Pool
}

// SetCommand method set the Command field of PoolAttach object.
func (p *PoolAttach) SetCommand(Command string) {
	p.Command = Command
}

// GetProperty method get the Property field of PoolAttach object.
func (p *PoolAttach) GetProperty() []string {
	return p.Property
}

// GetForcefully method get the Forcefully field of PoolAttach object.
func (p *PoolAttach) GetForcefully() bool {
	return p.Forcefully
}

// GetDevice method get the Device field of PoolAttach object.
func (p *PoolAttach) GetDevice() string {
	return p.Device
}

// GetNewDevice method get the NewDevice field of PoolAttach object.
func (p *PoolAttach) GetNewDevice() string {
	return p.NewDevice
}

// GetPool method get the Pool field of PoolAttach object.
func (p *PoolAttach) GetPool() string {
	return p.Pool
}

// GetCommand method get the Command field of PoolAttach object.
func (p *PoolAttach) GetCommand() string {
	return p.Command
}
