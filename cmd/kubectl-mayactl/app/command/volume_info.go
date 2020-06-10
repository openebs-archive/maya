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
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	client "github.com/openebs/maya/pkg/client/jiva"
	k8sclient "github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	v1 "github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumeInfoCommandHelpText = `
This command fetches information and status of the various
aspects of a Volume such as ISCSI, Controller, and Replica.

Usage: mayactl volume describe --volname <vol>
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
	IQN            string
	VolumeName     string
	Portal         string
	Size           string
	Status         string
	ReplicaCount   string
	ControllerNode string
}

// ReplicaInfo keep info about the replicas.
type ReplicaInfo struct {
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
		Use:     "describe",
		Short:   "Displays Openebs Volume information",
		Long:    volumeInfoCommandHelpText,
		Example: `mayactl volume describe --volname <vol>`,
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
	volumeInfo := &VolumeInfo{}
	// FetchVolumeInfo is called to get the volume controller's info such as
	// controller's IP, status, iqn, replica IPs etc.
	volumeInfo, err := NewVolumeInfo(mapiserver.GetURL()+VolumeAPIPath+c.volName, c.volName, c.namespace)
	if err != nil {
		return nil
	}

	// Initiallize an instance of ReplicaCollection, json response received from the replica controller. Collection contains status and other information of replica.
	collection := client.ReplicaCollection{}
	if volumeInfo.GetCASType() == string(JivaStorageEngine) {
		collection, err = getReplicaInfo(volumeInfo)
	}
	c.DisplayVolumeInfo(volumeInfo, collection)
	return nil
}

// getReplicaInfo returns the collection of replicas available for jiva volumes
func getReplicaInfo(volumeInfo *VolumeInfo) (client.ReplicaCollection, error) {
	controllerClient := client.ControllerClient{}
	collection := client.ReplicaCollection{}
	controllerStatuses := strings.Split(volumeInfo.GetControllerStatus(), ",")
	// Iterating over controllerStatus
	for _, controllerStatus := range controllerStatuses {
		if controllerStatus != controllerStatusOk {
			fmt.Printf("Unable to fetch volume details, Volume controller's status is '%s'.\n", controllerStatus)
			return collection, errors.New("Unable to fetch volume details")
		}
	}
	// controllerIP:9501/v1/replicas is to be parsed into this structure via GetVolumeStats.
	// An API needs to be passed as argument.
	_, err := controllerClient.GetVolumeStats(volumeInfo.GetClusterIP()+v1.ControllerPort, v1.InfoAPI, &collection)
	if err != nil {
		fmt.Printf("Cannot get volume stats %v", err)
	}
	return collection, err
}

// updateReplicaInfo parses replica information to replicaInfo structure
func updateReplicasInfo(replicaInfo map[int]*ReplicaInfo) error {
	K8sClient, err := k8sclient.NewK8sClient("")
	if err != nil {
		return err
	}

	pods, err := K8sClient.GetPods()
	if err != nil {
		return err
	}

	for _, replica := range replicaInfo {
		for _, pod := range pods {
			if pod.Status.PodIP == replica.IP {
				replica.NodeName = pod.Spec.NodeName
				replica.Name = pod.ObjectMeta.Name
			}
		}
	}

	return nil
}

// DisplayVolumeInfo displays the outputs in standard I/O.
// Currently it displays volume access modes and target portal details only.
func (c *CmdVolumeOptions) DisplayVolumeInfo(v *VolumeInfo, collection client.ReplicaCollection) error {
	var (
		// address and mode are used here as blackbox for the replica info
		// address keeps the ip and access mode details respectively.
		address, mode []string
		replicaCount  int
		portalInfo    PortalInfo
	)
	const (
		jivaReplicaTemplate = `
Replica Details :
-----------------
{{ printf "NAME\t ACCESSMODE\t STATUS\t IP\t NODE" }}
{{ printf "-----\t -----------\t -------\t ---\t -----" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.AccessMode }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.IP }} {{ $value.NodeName }} {{end}}
`

		cstorReplicaTemplate = `
Replica Details :
-----------------
{{ printf "%s\t" "NAME"}} {{ printf "%s\t" "STATUS"}} {{ printf "%s\t" "POOL NAME"}} {{ printf "%s\t" "NODE"}}
{{ printf "----\t ------\t ---------\t -----" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.PoolName }} {{ $value.NodeName }} {{end}}
`

		portalTemplate = `
Portal Details :
----------------
IQN               :   {{.IQN}}
Volume            :   {{.VolumeName}}
Portal            :   {{.Portal}}
Size              :   {{.Size}}
Controller Status :   {{.Status}}
Controller Node   :   {{.ControllerNode}}
Replica Count     :   {{.ReplicaCount}}
`
	)

	portalInfo = PortalInfo{
		v.GetIQN(),
		v.GetVolumeName(),
		v.GetTargetPortal(),
		v.GetVolumeSize(),
		v.GetControllerStatus(),
		v.GetReplicaCount(),
		v.GetControllerNode(),
	}

	tmpl, err := template.New("VolumeInfo").Parse(portalTemplate)
	if err != nil {
		fmt.Println("Error displaying output, found error :", err)
		return nil
	}
	err = tmpl.Execute(os.Stdout, portalInfo)
	if err != nil {
		fmt.Println("Error displaying volume details, found error :", err)
		return nil
	}

	if v.GetCASType() == string(JivaStorageEngine) {
		replicaCount, _ = strconv.Atoi(v.GetReplicaCount())
		// This case will occur only if user has manually specified zero replica.
		if replicaCount == 0 || len(v.GetReplicaStatus()) == 0 {
			fmt.Println("None of the replicas are running, please check the volume pod's status by running [kubectl describe pod -l=openebs/replica --all-namespaces] or try again later.")
			return nil
		}
		// Splitting strings with delimiter ','
		replicaStatusStrings := strings.Split(v.GetReplicaStatus(), ",")
		addressIPStrings := strings.Split(v.GetReplicaIP(), ",")

		// making a map of replica ip and their respective status,index and mode
		replicaIPStatus := make(map[string]*Value)

		// Creating a map of address and mode. The IP is chosen as key so that the status of that corresponding replica can be merged in linear time complexity
		for index, IP := range addressIPStrings {
			if strings.Contains(IP, "nil") {
				// appending address with index to avoid same key conflict as the IP is returned as `nil` in case of error
				replicaIPStatus[IP+string(index)] = &Value{index: index, status: replicaStatusStrings[index], mode: "NA"}
			} else {
				replicaIPStatus[IP] = &Value{index: index, status: replicaStatusStrings[index], mode: "NA"}
			}
		}

		// We get the info of the running replicas from the collection.data.
		// We are appending modes if available in collection.data to replicaIPStatus
		replicaInfo := make(map[int]*ReplicaInfo)

		for key := range collection.Data {
			address = append(address, strings.TrimSuffix(strings.TrimPrefix(collection.Data[key].Address, "tcp://"), v1.ReplicaPort))
			mode = append(mode, collection.Data[key].Mode)
			if _, ok := replicaIPStatus[address[key]]; ok {
				replicaIPStatus[address[key]].mode = mode[key]
			}
		}

		for IP, replicaStatus := range replicaIPStatus {
			// checking if the first three letters is nil or not if it is nil then the ip is not available
			if strings.Contains(IP, "nil") {
				replicaInfo[replicaStatus.index] = &ReplicaInfo{"NA", replicaStatus.mode, replicaStatus.status, "NA", "NA"}
			} else {
				replicaInfo[replicaStatus.index] = &ReplicaInfo{IP, replicaStatus.mode, replicaStatus.status, "NA", "NA"}
			}
		}
		// updating the replica info to replica structure
		err = updateReplicasInfo(replicaInfo)
		if err != nil {
			fmt.Println("Error in getting specific information from K8s. Please try again.")
		}

		return mapiserver.Print(jivaReplicaTemplate, replicaInfo)
	} else if v.GetCASType() == string(CstorStorageEngine) {

		// Converting replica count character to int
		replicaCount, err = strconv.Atoi(v.GetReplicaCount())
		if err != nil {
			fmt.Println("Invalid replica count")
			return nil
		}

		// Spitting the replica status
		replicaStatus := strings.Split(v.GetControllerStatus(), ",")
		poolName := strings.Split(v.GetStoragePool(), ",")
		cvrName := strings.Split(v.GetCVRName(), ",")
		nodeName := strings.Split(v.GetNodeName(), ",")

		// Confirming replica status, poolname , cvrName, nodeName are equal to replica count
		//if replicaCount != len(replicaStatus) || replicaCount != len(poolName) || replicaCount != len(cvrName) || replicaCount != len(nodeName) {
		if replicaCount != len(poolName) || replicaCount != len(cvrName) || replicaCount != len(nodeName) {
			fmt.Println("Invalid response received from maya-api service")
			return nil
		}

		replicaInfo := []cstorReplicaInfo{}

		// Iterating over the values replica values and appending to the structure
		for i := 0; i < replicaCount; i++ {
			replicaInfo = append(replicaInfo, cstorReplicaInfo{
				Name:       cvrName[i],
				PoolName:   poolName[i],
				AccessMode: "N/A",
				Status:     strings.Title(replicaStatus[i]),
				NodeName:   nodeName[i],
				IP:         "N/A",
			})
		}

		return mapiserver.Print(cstorReplicaTemplate, replicaInfo)
	} else {
		fmt.Println("Unsupported Volume Type")
	}
	return nil
}
