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

package pdestroy

// SetPool method set the Pool field of PoolDestroy object.
func (p *PoolDestroy) SetPool(Pool string) {
	p.Pool = Pool
}

// SetForcefully method set the Forcefully field of PoolDestroy object.
func (p *PoolDestroy) SetForcefully(Forcefully bool) {
	p.Forcefully = Forcefully
}

// SetCommand method set the Command field of PoolDestroy object.
func (p *PoolDestroy) SetCommand(Command string) {
	p.Command = Command
}

// GetPool method get the Pool field of PoolDestroy object.
func (p *PoolDestroy) GetPool() string {
	return p.Pool
}

// GetForcefully method get the Forcefully field of PoolDestroy object.
func (p *PoolDestroy) GetForcefully() bool {
	return p.Forcefully
}

// GetCommand method get the Command field of PoolDestroy object.
func (p *PoolDestroy) GetCommand() string {
	return p.Command
}
