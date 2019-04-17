/*
Copyright 2017 The Kubernetes Authors.

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

package driver

import (
	"encoding/json"
	"strings"

	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
	"k8s.io/utils/keymutex"
)

func portalMounter(portal string) string {
	if !strings.Contains(portal, ":") {
		portal = portal + ":3260"
	}
	return portal
}

func parseSecret(secretParams string) map[string]string {
	var secret map[string]string
	if err := json.Unmarshal([]byte(secretParams), &secret); err != nil {
		return nil
	}
	return secret
}

type iscsiDisk struct {
	Portals       []string
	Iqn           string
	lun           string
	Iface         string
	chapDiscovery bool
	chapSession   bool
	secret        map[string]string
	InitiatorName string
	VolName       string
}

type iscsiPlugin struct {
	host        volume.VolumeHost
	targetLocks keymutex.KeyMutex
}

type iscsiDiskMounter struct {
	*iscsiDisk
	readOnly     bool
	fsType       string
	mountOptions []string
	mounter      *mount.SafeFormatAndMount
	exec         mount.Exec
	deviceUtil   util.DeviceUtil
	targetPath   string
}

type iscsiDiskUnmounter struct {
	*iscsiDisk
	mounter mount.Interface
	exec    mount.Exec
}
