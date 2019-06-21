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

package premove

// SetPool method set the Pool field of PoolRemove object.
func (p *PoolRemove) SetPool(Pool string) {
	p.Pool = Pool
}

// SetDevice method set the Device field of PoolRemove object.
func (p *PoolRemove) SetDevice(Device string) {
	p.Device = append(p.Device, Device)
}

// SetCommand method set the Command field of PoolRemove object.
func (p *PoolRemove) SetCommand(Command string) {
	p.Command = Command
}

// GetPool method get the Pool field of PoolRemove object.
func (p *PoolRemove) GetPool() string {
	return p.Pool
}

// GetDevice method get the Device field of PoolRemove object.
func (p *PoolRemove) GetDevice() []string {
	return p.Device
}

// GetCommand method get the Command field of PoolRemove object.
func (p *PoolRemove) GetCommand() string {
	return p.Command
}
