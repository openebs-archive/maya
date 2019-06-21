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

package pexport

// SetAllPool method set the AllPool field of PoolExport object.
func (p *PoolExport) SetAllPool(AllPool bool) {
	p.AllPool = AllPool
}

// SetForcefully method set the Forcefully field of PoolExport object.
func (p *PoolExport) SetForcefully(Forcefully bool) {
	p.Forcefully = Forcefully
}

// SetPoolList method append the pool to PoolList field of PoolExport object.
func (p *PoolExport) SetPoolList(pool string) {
	p.PoolList = append(p.PoolList, pool)
}

// SetCommand method set the Command field of PoolExport object.
func (p *PoolExport) SetCommand(Command string) {
	p.Command = Command
}

// GetAllPool method get the AllPool field of PoolExport object.
func (p *PoolExport) GetAllPool() bool {
	return p.AllPool
}

// GetForcefully method get the Forcefully field of PoolExport object.
func (p *PoolExport) GetForcefully() bool {
	return p.Forcefully
}

// GetPoolList method get the PoolList field of PoolExport object.
func (p *PoolExport) GetPoolList() []string {
	return p.PoolList
}

// GetCommand method get the Command field of PoolExport object.
func (p *PoolExport) GetCommand() string {
	return p.Command
}
