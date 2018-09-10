/*
Copyright 2017 The OpenEBS Authors.

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

package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/spf13/cobra"
)

var (
	volumeInfoCommandHelpText = `
This command fetches information and status of the various
aspects of a Volume such as ISCSI, Controller, and Replica.

Usage: mayactl volume info --volname <vol>
`
)

// Value keeps info of the values of a current address in replicaIPStatus map
type Value struct {
	index  int
	status string
	mode   string
}

// PortalInfo keep info about the ISCSI Target Portal.
type PortalInfo struct {
	IQN          string
	VolumeName   string
	Portal       string
	Size         string
	Status       string
	ReplicaCount string
}

// ReplicaInfo keep info about the replicas.
type jivaReplicaInfo struct {
	IP         string
	AccessMode string
	Status     string
	Name       string
	NodeName   string
}

// cstorReplicaInfo holds information about the cstor replicas
type cstorReplicaInfo struct {
	Name       string
	PoolName   string
	AccessMode string
	Status     string
	NodeName   string
	IP         string
}

// NewCmdVolumeInfo displays OpenEBS Volume information.
func NewCmdVolumeInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Displays Openebs Volume information",
		Long:    volumeInfoCommandHelpText,
		Example: `mayactl volume info --volname <vol>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd, false, false, true), util.Fatal)
			util.CheckErr(options.RunVolumeInfo(cmd), util.Fatal)
		},
	}
	cmd.Flags().StringVarP(&options.volName, "volname", "", options.volName,
		"a unique volume name.")
	return cmd
}

// RunVolumeInfo runs info command and make call to DisplayVolumeInfo to display the results
func (c *CmdVolumeOptions) RunVolumeInfo(cmd *cobra.Command) error {
	// ReadVolume is called to get the volume controller's info such as
	// controller's IP, status, iqn, replica IPs etc.
	volume, err := mapiserver.ReadVolume(c.volName, c.namespace)
	if err != nil {
		CheckError(err)
	}

	volume, err = processCASVolume(volume, true)
	if err != nil {
		CheckError(err)
	}

	err = displayPortal(volume)
	if err != nil {
		CheckError(err)
	}

	if volume.Spec.CasType == string(JivaStorageEngine) {
		err = displayJivaReplicaDetails(volume)
	} else if volume.Spec.CasType == string(CstorStorageEngine) {
		err = displayCstorReplicaDetails(volume)
	} else {
		err = fmt.Errorf("Unsupported CASType found")
	}

	if err != nil {
		CheckError(err)
	}

	return nil
}

func displayPortal(v v1alpha1.CASVolume) error {
	const portalTemplate = `
Portal Details :
----------------
IQN           :   {{.IQN}}
Volume        :   {{.VolumeName}}
Portal        :   {{.Portal}}
Size          :   {{.Size}}
Status        :   {{.Status}}
Replica Count :   {{.ReplicaCount}}

	`

	portalInfo := PortalInfo{
		v.Spec.Iqn,
		v.ObjectMeta.Name,
		v.Spec.TargetPortal,
		v.Spec.Capacity,
		v.ObjectMeta.Annotations[controllerStatus],
		v.Spec.Replicas,
	}
	err := renderTemplate("VolumePortal", portalTemplate, portalInfo)
	return err
}

func displayJivaReplicaDetails(v v1alpha1.CASVolume) error {
	const jivaReplicaTemplate = `
Replica Details :
-----------------
{{ printf "NAME\t ACCESSMODE\t STATUS\t IP\t NODE" }}
{{ printf "-----\t -----------\t -------\t ---\t -----" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.AccessMode }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.IP }} {{ $value.NodeName }} {{end}}

`

	// Convert replica count character to int
	jivaReplicaCount, err := strconv.Atoi(v.Spec.Replicas)
	if err != nil {
		CheckError(fmt.Errorf("Invalid replica count"))
	}
	// Splitting values seperated by delemiter
	jivaReplicaNodeName := strings.Split(v.ObjectMeta.Annotations[replicaNodeName], ",")
	jivaReplicaPodName := strings.Split(v.ObjectMeta.Annotations[replicaPodName], ",")
	jivaReplicaAccessMode := strings.Split(v.ObjectMeta.Annotations[replicaAccessMode], ",")
	jivaReplicaIP := strings.Split(v.ObjectMeta.Annotations[replicaIP], ",")
	jivaReplicaStatus := strings.Split(v.ObjectMeta.Annotations[replicaStatus], ",")
	jivaReplicas := []jivaReplicaInfo{}

	// Confirm replica status, podname , accessmode, nodeName and IP are equal to replica count
	checkInvalidResponse := len(jivaReplicaAccessMode) != jivaReplicaCount || len(jivaReplicaIP) != jivaReplicaCount || len(jivaReplicaNodeName) != jivaReplicaCount || len(jivaReplicaPodName) != jivaReplicaCount || len(jivaReplicaStatus) != jivaReplicaCount
	if checkInvalidResponse {
		return fmt.Errorf("Invalid replica response received from maya api service")
	}

	// Iterating over the values replica values and appending to the structure
	for index := 0; index < jivaReplicaCount; index++ {
		jivaReplicas = append(jivaReplicas, jivaReplicaInfo{
			Name:       jivaReplicaPodName[index],
			NodeName:   jivaReplicaNodeName[index],
			AccessMode: jivaReplicaAccessMode[index],
			Status:     jivaReplicaStatus[index],
			IP:         jivaReplicaIP[index],
		})
	}

	err = renderTemplate("JivaReplicaInfo", jivaReplicaTemplate, jivaReplicas)
	if err != nil {
		return err
	}
	return nil
}

func displayCstorReplicaDetails(v v1alpha1.CASVolume) error {
	const cstorReplicaTemplate = `
Replica Details :
-----------------
{{ printf "%s\t" "NAME"}} {{ printf "%s\t" "STATUS"}} {{ printf "%s\t" "POOL NAME"}} {{ printf "%s\t" "NODE"}}
{{ printf "----\t ------\t ---------\t -----" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.PoolName }} {{ $value.NodeName }} {{end}}
	
`
	// Convert replica count character to int
	cstorReplicaCount, err := strconv.Atoi(v.Spec.Replicas)
	if err != nil {
		fmt.Println("Invalid replica count")
		return nil
	}

	// Split values seperated by delemiter
	cstorControllerStatus := strings.Split(v.ObjectMeta.Annotations[controllerStatus], ",")
	cstorPoolName := strings.Split(v.ObjectMeta.Annotations[storagePool], ",")
	cstorCVRName := strings.Split(v.ObjectMeta.Annotations[cvrName], ",")
	cstorNodeName := strings.Split(v.ObjectMeta.Annotations[nodeName], ",")
	cstorReplicas := []cstorReplicaInfo{}

	// Confirm replica status, poolname , cvrName and nodeName are equal to replica count
	checkInvalidResponse := cstorReplicaCount != len(cstorPoolName) || cstorReplicaCount != len(cstorCVRName) || cstorReplicaCount != len(cstorNodeName) || cstorReplicaCount >= len(cstorControllerStatus)
	if checkInvalidResponse {
		return fmt.Errorf("Invalid replica response received from maya api service")
	}

	// Iterating over the values replica values and appending to the structure
	for index := 0; index < cstorReplicaCount; index++ {
		cstorReplicas = append(cstorReplicas, cstorReplicaInfo{
			Name:       cstorCVRName[index],
			PoolName:   cstorPoolName[index],
			AccessMode: "N/A",
			Status:     cstorControllerStatus[index],
			NodeName:   cstorNodeName[index],
			IP:         "N/A",
		})
	}

	err = renderTemplate("CstorReplicaInfo", cstorReplicaTemplate, cstorReplicas)
	if err != nil {
		return err
	}

	return nil
}
