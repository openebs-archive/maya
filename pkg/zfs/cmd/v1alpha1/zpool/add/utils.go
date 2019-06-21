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

// SetVdev method set the VdevList field of PoolExpansion object.
func (p *PoolExpansion) SetVdev(Vdev string) {
	p.VdevList = append(p.VdevList, Vdev)
}

// SetProperty method set the Property field of PoolExpansion object.
func (p *PoolExpansion) SetProperty(key, value string) {
	p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetPool method set the Pool field of PoolExpansion object.
func (p *PoolExpansion) SetPool(Pool string) {
	p.Pool = Pool
}

// SetCommand method set the Command field of PoolExpansion object.
func (p *PoolExpansion) SetCommand(Command string) {
	p.Command = Command
}

// GetVdevList method get the VdevList field of PoolExpansion object.
func (p *PoolExpansion) GetVdevList() []string {
	return p.VdevList
}

// GetProperty method get the Property field of PoolExpansion object.
func (p *PoolExpansion) GetProperty() []string {
	return p.Property
}

// GetPool method get the Pool field of PoolExpansion object.
func (p *PoolExpansion) GetPool() string {
	return p.Pool
}

// GetCommand method get the Command field of PoolExpansion object.
func (p *PoolExpansion) GetCommand() string {
	return p.Command
}
