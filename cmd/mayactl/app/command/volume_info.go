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
	"text/tabwriter"

	"github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/jiva"
	k8sclient "github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/pkg/client/mapiserver"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumeInfoCommandHelpText = `
This command fetches information and status of the various
aspects of a Volume such as ISCSI, Controller, and Replica.

Usage: mayactl volume info --volname <vol>
`
)

const (
	listPath = "/latest/volumes/"
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
type ReplicaInfo struct {
	IP         string
	AccessMode string
	Status     string
	Name       string
	NodeName   string
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

// RunVolumeInfo runs info command and make call to DisplayVolumeInfo
func (c *CmdVolumeOptions) RunVolumeInfo(cmd *cobra.Command) error {
	volumeInfo := &v1alpha1.CASVolume{}
	// FetchVolumeInfo is called to get the volume controller's info such as
	// controller's IP, status, iqn, replica IPs etc.
	err := volumeInfo.FetchVolumeInfo(mapiserver.GetURL()+listPath+c.volName, c.volName, c.namespace)
	if err != nil {
		return err
	}

	// Initiallize an instance of ReplicaCollection, json response recieved from the
	collection := client.ReplicaCollection{}
	if volumeInfo.GetField("CasType") == "jiva" {
		collection, err = getReplicaInfo(volumeInfo)
	}
	c.DisplayVolumeInfo(volumeInfo, collection)
	return nil
}

func getReplicaInfo(volumeInfo *v1alpha1.CASVolume) (client.ReplicaCollection, error) {
	controllerClient, collection, controllerStatuses := client.ControllerClient{}, client.ReplicaCollection{}, strings.Split(volumeInfo.GetField("ControllerStatus"), ",")
	for _, controllerStatus := range controllerStatuses {
		if controllerStatus != "running" {
			fmt.Printf("Unable to fetch volume details, Volume controller's status is '%s'.\n", controllerStatus)
			return collection, errors.New("Unable to fetch volume details")
		}
	}
	// controllerIP:9501/v1/replicas is to be parsed into this structure via GetVolumeStats.
	// An API needs to be passed as argument.
	_, err := controllerClient.GetVolumeStats(volumeInfo.GetField("ClusterIP")+v1.ControllerPort, v1.InfoAPI, &collection)
	if err != nil {
		fmt.Printf("Cannot get volume stats %v", err)
	}
	return collection, err
}

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
func (c *CmdVolumeOptions) DisplayVolumeInfo(v *v1alpha1.CASVolume, collection client.ReplicaCollection) error {
	var (
		// address and mode are used here as blackbox for the replica info
		// address keeps the ip and access mode details respectively.
		address, mode []string
		replicaCount  int
		portalInfo    PortalInfo
	)
	const (
		replicaTemplate = `

Replica Details :
----------------
{{ printf "NAME\t ACCESSMODE\t STATUS\t IP\t NODE" }}
{{ printf "-----\t -----------\t -------\t ---\t -----" }} {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.AccessMode }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.IP }} {{ $value.NodeName }} {{end}}
`
		portalTemplate = `
Portal Details :
---------------
IQN           :   {{.IQN}}
Volume        :   {{.VolumeName}}
Portal        :   {{.Portal}}
Size          :   {{.Size}}
Status        :   {{.Status}}
Replica Count :   {{.ReplicaCount}}
`
	)

	portalInfo = PortalInfo{
		v.GetField("IQN"),
		v.GetField("VolumeName"),
		v.GetField("TargetPortal"),
		v.GetField("VolumeSize"),
		v.GetField("ControllerStatus"),
		v.GetField("ReplicaCount"),
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
	if v.GetField("CasType") == "jiva" {
		replicaCount, _ = strconv.Atoi(v.GetField("ReplicaCount"))
		// This case will occur only if user has manually specified zero replica.
		if replicaCount == 0 || v.GetField("ReplicaStatus") == "" {
			fmt.Println("None of the replicas are running, please check the volume pod's status by running [kubectl describe pod -l=openebs/replica --all-namespaces] or try again later.")
			return nil
		}
		// Splitting strings with delimiter ','
		replicaStatusStrings := strings.Split(v.GetField("ReplicaStatus"), ",")
		addressIPStrings := strings.Split(v.GetField("ReplicaIP"), ",")

		// making a map of replica ip and their respective status,index and mode
		replicaIPStatus := make(map[string]*Value)

		// Creating a map of address and mode
		for index, IP := range addressIPStrings {
			if IP != "nil" {
				replicaIPStatus[IP] = &Value{index: index, status: replicaStatusStrings[index], mode: "NA"}
			} else {
				// appending address with index to avoid same key conflict
				replicaIPStatus[IP+string(index)] = &Value{index: index, status: replicaStatusStrings[index], mode: "NA"}
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
			// checking if the first three letters is nil or not if it is nil then the ip is not avaiable
			if IP[0:3] != "nil" {
				replicaInfo[replicaStatus.index] = &ReplicaInfo{IP, replicaStatus.mode, replicaStatus.status, "NA", "NA"}
			} else {
				replicaInfo[replicaStatus.index] = &ReplicaInfo{"NA", replicaStatus.mode, replicaStatus.status, "NA", "NA"}
			}
		}

		err = updateReplicasInfo(replicaInfo)
		if err != nil {
			fmt.Println("Error in getting specific information from K8s. Please try again.")
		}

		tmpl = template.New("ReplicaInfo")
		tmpl = template.Must(tmpl.Parse(replicaTemplate))

		w := tabwriter.NewWriter(os.Stdout, v1.MinWidth, v1.MaxWidth, v1.Padding, ' ', 0)
		err = tmpl.Execute(w, replicaInfo)
		if err != nil {
			fmt.Println("Unable to display volume info, found error : ", err)
		}
		w.Flush()
	}
	return nil
}
