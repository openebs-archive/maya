// +build !ignore_autogenerated

/*
Copyright 2017 The OpenEBS Authors

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	reflect "reflect"

	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
//
// Deprecated: deepcopy registration will go away when static deepcopy is fully implemented.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorPool).DeepCopyInto(out.(*CstorPool))
			return nil
		}, InType: reflect.TypeOf(&CstorPool{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorPoolAttr).DeepCopyInto(out.(*CstorPoolAttr))
			return nil
		}, InType: reflect.TypeOf(&CstorPoolAttr{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorPoolList).DeepCopyInto(out.(*CstorPoolList))
			return nil
		}, InType: reflect.TypeOf(&CstorPoolList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorPoolSpec).DeepCopyInto(out.(*CstorPoolSpec))
			return nil
		}, InType: reflect.TypeOf(&CstorPoolSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorReplica).DeepCopyInto(out.(*CstorReplica))
			return nil
		}, InType: reflect.TypeOf(&CstorReplica{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorReplicaList).DeepCopyInto(out.(*CstorReplicaList))
			return nil
		}, InType: reflect.TypeOf(&CstorReplicaList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*CstorReplicaSpec).DeepCopyInto(out.(*CstorReplicaSpec))
			return nil
		}, InType: reflect.TypeOf(&CstorReplicaSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Policy).DeepCopyInto(out.(*Policy))
			return nil
		}, InType: reflect.TypeOf(&Policy{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*RunTasks).DeepCopyInto(out.(*RunTasks))
			return nil
		}, InType: reflect.TypeOf(&RunTasks{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StoragePool).DeepCopyInto(out.(*StoragePool))
			return nil
		}, InType: reflect.TypeOf(&StoragePool{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StoragePoolClaim).DeepCopyInto(out.(*StoragePoolClaim))
			return nil
		}, InType: reflect.TypeOf(&StoragePoolClaim{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StoragePoolClaimList).DeepCopyInto(out.(*StoragePoolClaimList))
			return nil
		}, InType: reflect.TypeOf(&StoragePoolClaimList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StoragePoolClaimSpec).DeepCopyInto(out.(*StoragePoolClaimSpec))
			return nil
		}, InType: reflect.TypeOf(&StoragePoolClaimSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StoragePoolList).DeepCopyInto(out.(*StoragePoolList))
			return nil
		}, InType: reflect.TypeOf(&StoragePoolList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*StoragePoolSpec).DeepCopyInto(out.(*StoragePoolSpec))
			return nil
		}, InType: reflect.TypeOf(&StoragePoolSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*Task).DeepCopyInto(out.(*Task))
			return nil
		}, InType: reflect.TypeOf(&Task{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*VolumePolicy).DeepCopyInto(out.(*VolumePolicy))
			return nil
		}, InType: reflect.TypeOf(&VolumePolicy{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*VolumePolicyList).DeepCopyInto(out.(*VolumePolicyList))
			return nil
		}, InType: reflect.TypeOf(&VolumePolicyList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*VolumePolicySpec).DeepCopyInto(out.(*VolumePolicySpec))
			return nil
		}, InType: reflect.TypeOf(&VolumePolicySpec{})},
	)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorPool) DeepCopyInto(out *CstorPool) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorPool.
func (in *CstorPool) DeepCopy() *CstorPool {
	if in == nil {
		return nil
	}
	out := new(CstorPool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CstorPool) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorPoolAttr) DeepCopyInto(out *CstorPoolAttr) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorPoolAttr.
func (in *CstorPoolAttr) DeepCopy() *CstorPoolAttr {
	if in == nil {
		return nil
	}
	out := new(CstorPoolAttr)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorPoolList) DeepCopyInto(out *CstorPoolList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CstorPool, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorPoolList.
func (in *CstorPoolList) DeepCopy() *CstorPoolList {
	if in == nil {
		return nil
	}
	out := new(CstorPoolList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CstorPoolList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorPoolSpec) DeepCopyInto(out *CstorPoolSpec) {
	*out = *in
	if in.Disks != nil {
		in, out := &in.Disks, &out.Disks
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.Poolspec = in.Poolspec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorPoolSpec.
func (in *CstorPoolSpec) DeepCopy() *CstorPoolSpec {
	if in == nil {
		return nil
	}
	out := new(CstorPoolSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorReplica) DeepCopyInto(out *CstorReplica) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorReplica.
func (in *CstorReplica) DeepCopy() *CstorReplica {
	if in == nil {
		return nil
	}
	out := new(CstorReplica)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CstorReplica) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorReplicaList) DeepCopyInto(out *CstorReplicaList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CstorReplica, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorReplicaList.
func (in *CstorReplicaList) DeepCopy() *CstorReplicaList {
	if in == nil {
		return nil
	}
	out := new(CstorReplicaList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CstorReplicaList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CstorReplicaSpec) DeepCopyInto(out *CstorReplicaSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CstorReplicaSpec.
func (in *CstorReplicaSpec) DeepCopy() *CstorReplicaSpec {
	if in == nil {
		return nil
	}
	out := new(CstorReplicaSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Policy) DeepCopyInto(out *Policy) {
	*out = *in
	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Policy.
func (in *Policy) DeepCopy() *Policy {
	if in == nil {
		return nil
	}
	out := new(Policy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RunTasks) DeepCopyInto(out *RunTasks) {
	*out = *in
	if in.Tasks != nil {
		in, out := &in.Tasks, &out.Tasks
		*out = make([]Task, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RunTasks.
func (in *RunTasks) DeepCopy() *RunTasks {
	if in == nil {
		return nil
	}
	out := new(RunTasks)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePool) DeepCopyInto(out *StoragePool) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePool.
func (in *StoragePool) DeepCopy() *StoragePool {
	if in == nil {
		return nil
	}
	out := new(StoragePool)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StoragePool) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaim) DeepCopyInto(out *StoragePoolClaim) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaim.
func (in *StoragePoolClaim) DeepCopy() *StoragePoolClaim {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StoragePoolClaim) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimList) DeepCopyInto(out *StoragePoolClaimList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StoragePoolClaim, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimList.
func (in *StoragePoolClaimList) DeepCopy() *StoragePoolClaimList {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StoragePoolClaimList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolClaimSpec) DeepCopyInto(out *StoragePoolClaimSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolClaimSpec.
func (in *StoragePoolClaimSpec) DeepCopy() *StoragePoolClaimSpec {
	if in == nil {
		return nil
	}
	out := new(StoragePoolClaimSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolList) DeepCopyInto(out *StoragePoolList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]StoragePool, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolList.
func (in *StoragePoolList) DeepCopy() *StoragePoolList {
	if in == nil {
		return nil
	}
	out := new(StoragePoolList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StoragePoolList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoragePoolSpec) DeepCopyInto(out *StoragePoolSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoragePoolSpec.
func (in *StoragePoolSpec) DeepCopy() *StoragePoolSpec {
	if in == nil {
		return nil
	}
	out := new(StoragePoolSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Task) DeepCopyInto(out *Task) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Task.
func (in *Task) DeepCopy() *Task {
	if in == nil {
		return nil
	}
	out := new(Task)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumePolicy) DeepCopyInto(out *VolumePolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumePolicy.
func (in *VolumePolicy) DeepCopy() *VolumePolicy {
	if in == nil {
		return nil
	}
	out := new(VolumePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumePolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumePolicyList) DeepCopyInto(out *VolumePolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]VolumePolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumePolicyList.
func (in *VolumePolicyList) DeepCopy() *VolumePolicyList {
	if in == nil {
		return nil
	}
	out := new(VolumePolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumePolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumePolicySpec) DeepCopyInto(out *VolumePolicySpec) {
	*out = *in
	if in.Policies != nil {
		in, out := &in.Policies, &out.Policies
		*out = make([]Policy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.RunTasks.DeepCopyInto(&out.RunTasks)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumePolicySpec.
func (in *VolumePolicySpec) DeepCopy() *VolumePolicySpec {
	if in == nil {
		return nil
	}
	out := new(VolumePolicySpec)
	in.DeepCopyInto(out)
	return out
}
