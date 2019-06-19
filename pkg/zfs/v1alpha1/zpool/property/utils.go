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

package zfs

import "fmt"

// SetPropList method set the PropList field of PoolProperty object.
func (p *PoolProperty) SetPropList(key, value string) {
	p.PropList = append(p.PropList, fmt.Sprintf("%s=%s", key, value))
}

// SetOpSet method set the OpSet field of PoolProperty object.
func (p *PoolProperty) SetOpSet(OpSet bool) {
	p.OpSet = OpSet
}

// SetPool method set the Pool field of PoolProperty object.
func (p *PoolProperty) SetPool(Pool string) {
	p.Pool = Pool
}

// SetCommand method set the Command field of PoolProperty object.
func (p *PoolProperty) SetCommand(Command string) {
	p.Command = Command
}

// GetPropList method get the PropList field of PoolProperty object.
func (p *PoolProperty) GetPropList() []string {
	return p.PropList
}

// GetOpSet method get the OpSet field of PoolProperty object.
func (p *PoolProperty) GetOpSet() bool {
	return p.OpSet
}

// GetPool method get the Pool field of PoolProperty object.
func (p *PoolProperty) GetPool() string {
	return p.Pool
}

// GetCommand method get the Command field of PoolProperty object.
func (p *PoolProperty) GetCommand() string {
	return p.Command
}
