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

package pcreate

import "fmt"

// SetProperty method append the Property to PoolCreate object.
func (p *PoolCreate) SetProperty(key, value string) {
	p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetPool method set the Pool field of PoolCreate object.
func (p *PoolCreate) SetPool(Pool string) {
	p.Pool = Pool
}

// SetVdev method set the Vdev field of PoolCreate object.
func (p *PoolCreate) SetVdev(Vdev string) {
	p.Vdev = append(p.Vdev, Vdev)
}

// SetForcefully method set the Forcefully field of PoolCreate object.
func (p *PoolCreate) SetForcefully(Forcefully bool) {
	p.Forcefully = Forcefully
}

// SetCommand method set the Command field of PoolCreate object.
func (p *PoolCreate) SetCommand(Command string) {
	p.Command = Command
}

// GetProperty method get the Property field of PoolCreate object.
func (p *PoolCreate) GetProperty() []string {
	return p.Property
}

// GetPool method get the Pool field of PoolCreate object.
func (p *PoolCreate) GetPool() string {
	return p.Pool
}

// GetVdev method get the Vdev field of PoolCreate object.
func (p *PoolCreate) GetVdev() []string {
	return p.Vdev
}

// GetForcefully method get the Forcefully field of PoolCreate object.
func (p *PoolCreate) GetForcefully() bool {
	return p.Forcefully
}

// GetCommand method get the Command field of PoolCreate object.
func (p *PoolCreate) GetCommand() string {
	return p.Command
}
