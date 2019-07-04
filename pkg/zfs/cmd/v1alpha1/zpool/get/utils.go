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

package pget

// SetPropList method set the PropList field of PoolGProperty object.
func (p *PoolGProperty) SetPropList(key string) {
	p.PropList = append(p.PropList, key)
}

// SetPool method set the Pool field of PoolGProperty object.
func (p *PoolGProperty) SetPool(Pool string) {
	p.Pool = Pool
}

// SetCommand method set the Command field of PoolGProperty object.
func (p *PoolGProperty) SetCommand(Command string) {
	p.Command = Command
}

// GetPropList method get the PropList field of PoolGProperty object.
func (p *PoolGProperty) GetPropList() []string {
	return p.PropList
}

// GetPool method get the Pool field of PoolGProperty object.
func (p *PoolGProperty) GetPool() string {
	return p.Pool
}

// GetCommand method get the Command field of PoolGProperty object.
func (p *PoolGProperty) GetCommand() string {
	return p.Command
}
