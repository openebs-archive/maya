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

package pimport

import "fmt"

// SetCachefile method set the Cachefile field of PoolImport object.
func (p *PoolImport) SetCachefile(Cachefile string) {
	p.Cachefile = Cachefile
}

// SetDirectorylist method append the directory to Directorylist field of PoolImport object.
func (p *PoolImport) SetDirectorylist(dir string) {
	p.Directorylist = append(p.Directorylist, dir)
}

// SetImportAll method set the ImportAll field of PoolImport object.
func (p *PoolImport) SetImportAll(ImportAll bool) {
	p.ImportAll = ImportAll
}

// SetForceImport method set the ForceImport field of PoolImport object.
func (p *PoolImport) SetForceImport(ForceImport bool) {
	p.ForceImport = ForceImport
}

// SetProperty method append the Propertyto field of PoolImport object property.
func (p *PoolImport) SetProperty(key, value string) {
	p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
}

// SetPool method set the Pool field of PoolImport object.
func (p *PoolImport) SetPool(Pool string) {
	p.Pool = Pool
}

// SetNewPool method set the NewPool field of PoolImport object.
func (p *PoolImport) SetNewPool(NewPool string) {
	p.NewPool = NewPool
}

// SetCommand method set the Command field of PoolImport object.
func (p *PoolImport) SetCommand(Command string) {
	p.Command = Command
}

// GetCachefile method get the Cachefile field of PoolImport object.
func (p *PoolImport) GetCachefile() string {
	return p.Cachefile
}

// GetDirectorylist method get the Directorylist field of PoolImport object.
func (p *PoolImport) GetDirectorylist() []string {
	return p.Directorylist
}

// GetImportAll method get the ImportAll field of PoolImport object.
func (p *PoolImport) GetImportAll() bool {
	return p.ImportAll
}

// GetForceImport method get the ForceImport field of PoolImport object.
func (p *PoolImport) GetForceImport() bool {
	return p.ForceImport
}

// GetProperty method get the Property field of PoolImport object.
func (p *PoolImport) GetProperty() []string {
	return p.Property
}

// GetPool method get the Pool field of PoolImport object.
func (p *PoolImport) GetPool() string {
	return p.Pool
}

// GetNewPool method get the NewPool field of PoolImport object.
func (p *PoolImport) GetNewPool() string {
	return p.NewPool
}

// GetCommand method get the Command field of PoolImport object.
func (p *PoolImport) GetCommand() string {
	return p.Command
}
