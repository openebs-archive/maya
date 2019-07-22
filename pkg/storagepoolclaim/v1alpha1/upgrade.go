/*
Copyright 2019 The OpenEBS Authors

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

package v1alpha1

import (
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PreUpgradeAction string

const (
	DisableReconciler	PreUpgradeAction = "DisableReconciler"
	Continue		PreUpgradeAction = "Continue"
	Abort			PreUpgradeAction = "Abort"
)

func (Spc *SPC) validVersionForPreUpgrade() {
}

func (Spc *SPC) requiresPreUpgrade() (PreUpgradeAction, error) {
	if Spc.HasFinalizer(SPCFinalizer) {
		return Abort, nil
	}
	return Continue, nil
}

type preUpgradeFn func(*SPC) error

var performPreUpgradeFn = map[PreUpgradeAction]preUpgradeFn {
	DisableReconciler:	noop,
	Continue:		performPreUpgrade,
	Abort:			noop,
}

func noop(*SPC) error {
	return nil
}

func (Spc *SPC) performPreUpgradeOnAssociatedBDCs() error {
	return Spc.addSPCFinalizerOnAssociatedBDCs()
}

func performPreUpgrade(Spc *SPC) error {
	err := Spc.performPreUpgradeOnAssociatedBDCs()
	if (err != nil) {
		return err
	}
	_, err = Spc.AddFinalizer(SPCFinalizer)
	return err
}

func (Spc *SPC) preUpgrade() error {
	res, err := Spc.requiresPreUpgrade()
	if err != nil {
		return err
	}
	err = performPreUpgradeFn[res](Spc)
	return err
}

func PreUpgrade() error {
	spcList, _ := NewKubeClient().List(metav1.ListOptions {})
	for _, obj := range spcList.Items {
		obj := obj
		Spc := SPC{&obj}
		Spc.preUpgrade()
	}
	return nil
}
