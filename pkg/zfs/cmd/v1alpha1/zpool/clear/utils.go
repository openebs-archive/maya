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

package pclear

// SetPool method set the Pool field of PoolClear object.
func (p *PoolClear) SetPool(Pool string) {
	p.Pool = Pool
}

// SetVdev method set the Vdev field of PoolClear object.
func (p *PoolClear) SetVdev(Vdev string) {
	p.Vdev = append(p.Vdev, Vdev)
}

// SetCommand method set the Command field of PoolClear object.
func (p *PoolClear) SetCommand(Command string) {
	p.Command = Command
}

// GetPool method get the Pool field of PoolClear object.
func (p *PoolClear) GetPool() string {
	return p.Pool
}

// GetVdev method get the Vdev field of PoolClear object.
func (p *PoolClear) GetVdev() []string {
	return p.Vdev
}

// GetCommand method get the Command field of PoolClear object.
func (p *PoolClear) GetCommand() string {
	return p.Command
}
