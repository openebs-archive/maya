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

package vrollback

// SetDestroy method set the Destroy field of VolumeRollback object.
func (v *VolumeRollback) SetDestroy(Destroy bool) {
	v.Destroy = Destroy
}

// SetForceUnmount method set the ForceUnmount field of VolumeRollback object.
func (v *VolumeRollback) SetForceUnmount(ForceUnmount bool) {
	v.ForceUnmount = ForceUnmount
}

// SetDestroySnap method set the DestroySnap field of VolumeRollback object.
func (v *VolumeRollback) SetDestroySnap(DestroySnap bool) {
	v.DestroySnap = DestroySnap
}

// SetSnapshot method set the Snapshot field of VolumeRollback object.
func (v *VolumeRollback) SetSnapshot(Snapshot string) {
	v.Snapshot = Snapshot
}

// SetCommand method set the Command field of VolumeRollback object.
func (v *VolumeRollback) SetCommand(Command string) {
	v.Command = Command
}

// GetDestroy method get the Destroy field of VolumeRollback object.
func (v *VolumeRollback) GetDestroy() bool {
	return v.Destroy
}

// GetForceUnmount method get the ForceUnmount field of VolumeRollback object.
func (v *VolumeRollback) GetForceUnmount() bool {
	return v.ForceUnmount
}

// GetDestroySnap method get the DestroySnap field of VolumeRollback object.
func (v *VolumeRollback) GetDestroySnap() bool {
	return v.DestroySnap
}

// GetSnapshot method get the Snapshot field of VolumeRollback object.
func (v *VolumeRollback) GetSnapshot() string {
	return v.Snapshot
}

// GetCommand method get the Command field of VolumeRollback object.
func (v *VolumeRollback) GetCommand() string {
	return v.Command
}
