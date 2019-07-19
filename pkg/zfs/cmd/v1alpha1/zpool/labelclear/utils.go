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

package plabelclear

// SetVdev method set the Vdev field of PoolLabelClear object.
func (p *PoolLabelClear) SetVdev(Vdev string) {
	p.Vdev = Vdev
}

// SetForcefully method set the Forcefully field of PoolLabelClear object.
func (p *PoolLabelClear) SetForcefully(Forcefully bool) {
	p.Forcefully = Forcefully
}

// SetCommand method set the Command field of PoolLabelClear object.
func (p *PoolLabelClear) SetCommand(Command string) {
	p.Command = Command
}
