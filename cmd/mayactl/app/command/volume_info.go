package command

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	client "github.com/openebs/maya/pkg/client/jiva"
	k8sclient "github.com/openebs/maya/pkg/client/k8s"
	"github.com/openebs/maya/pkg/util"
	"github.com/openebs/maya/types/v1"
	"github.com/spf13/cobra"
)

var (
	volumeInfoCommandHelpText = `
	    Usage: mayactl volume info --volname <vol>

        This command fetches the information and status of the various
	    aspects of the Volume such as ISCSI, Controller and Replica.
        `
)

//values keeps info of the values of a current address in replicaIPStatus map
type Value struct {
	index  int
	status string
	mode   string
}

// PortalInfo keep info about the ISCSI Target Portal.
type PortalInfo struct {
	IQN        string
	VolumeName string
	Portal     string
	Size       string
	Status     string
}

// ReplicaInfo keep info about the replicas.
type ReplicaInfo struct {
	IP         string
	AccessMode string
	Status     string
	Name       string
	NodeName   string
}

// CmdVolumeInfoOptions is used to store the value of flags used in the cli
type CmdVolumeInfoOptions struct {
	volName string
}

// NewCmdVolumeInfo shows info of OpenEBS Volume
func NewCmdVolumeInfo() *cobra.Command {
	options := CmdVolumeInfoOptions{}
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Displays the info of Volume",
		Long:    volumeInfoCommandHelpText,
		Example: `mayactl volume info --volname <vol>`,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(options.Validate(cmd), util.Fatal)
			util.CheckErr(options.RunVolumeInfo(cmd), util.Fatal)
		},
	}
	cmd.Flags().StringVarP(&options.volName, "volname", "n", options.volName,
		"unique volume name.")

	return cmd
}

// Validate verifies the command whether volName is passed or not.
func (c *CmdVolumeInfoOptions) Validate(cmd *cobra.Command) error {
	if c.volName == "" {
		return errors.New("--volname is missing. Please try running [mayactl volume list] to see list of volumes.")
	}
	return nil
}

// TODO : Add more volume information
// RunVolumeInfo runs info command and make call to DisplayVolumeInfo
func (c *CmdVolumeInfoOptions) RunVolumeInfo(cmd *cobra.Command) error {
	annotation := &Annotations{}
	// GetVolumeAnnotation is called to get the volume controller's info such as
	// controller's IP, status, iqn, replica IPs etc.
	err := annotation.GetVolAnnotations(c.volName)
	if err != nil {
		return nil
	}
	if annotation.ControllerStatus != "Running" {
		fmt.Printf("Unable to fetch volume details, Volume controller's status is '%s'.\n", annotation.ControllerStatus)
		return nil
	}

	// Initiallize an instance of ReplicaCollection, json response recieved from the
	// controllerIP:9501/v1/replicas is to be parsed into this structure via GetVolumeStats.
	// An API needs to be passed as argument.
	collection := client.ReplicaCollection{}
	controllerClient := client.ControllerClient{}
	_, err = controllerClient.GetVolumeStats(annotation.ClusterIP+v1.ControllerPort, v1.InfoAPI, &collection)
	if err != nil {
		return err
	}

	c.DisplayVolumeInfo(annotation, collection)
	return nil
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
// Currently It displays volume access modes and target portal details only.
func (c *CmdVolumeInfoOptions) DisplayVolumeInfo(a *Annotations, collection client.ReplicaCollection) error {
	var (
		// address, mode are used here as blackbox for the replica info
		// address keeps the ip and access mode details respectively.
		address, mode []string
		replicaCount  int
		portalInfo    PortalInfo
	)
	const (
		replicaTemplate = `
<<<<<<< 47d72fc44065a7693c2b53a6ddd45a234114afdf
		
Replica Details : 
---------------- {{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.AccessMode }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.IP }} {{ $value.NodeName }} {{end}}
=======
============================================== Replica Details ==============================================
{{range $key, $value := .}}
{{ printf "%s\t" $value.Name }} {{ printf "%s\t" $value.AccessMode }} {{ printf "%s\t" $value.Status }} {{ printf "%s\t" $value.IP }} {{ $value.NodeName }} {{end}}
=============================================================================================================
>>>>>>> Proper alignment with the use of tabwriter
`
		portalTemplate = `
Portal Details : 
---------------
IQN     :   {{.IQN}}
Volume  :   {{.VolumeName}}
Portal  :   {{.Portal}}
Size    :   {{.Size}}
Status  :   {{.Status}}
`
	)
	portalInfo = PortalInfo{
		a.Iqn,
		c.volName,
		a.TargetPortal,
		a.VolSize,
		a.ControllerStatus,
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
	replicaCount, _ = strconv.Atoi(a.ReplicaCount)
	// This case will occur only if user has manually specified zero replica.
	if replicaCount == 0 {
		fmt.Println("None of the replicas are running, please check the volume pod's status by running [kubectl describe pod -l=openebs/replica --all-namespaces] or try again later.")
		return nil
	}

	// Splitting strings with delimiter ','
	replicaStatusStrings := strings.Split(a.ReplicaStatus, ",")
	addressIPStrings := strings.Split(a.Replicas, ",")

	// making a map of replica ip and their respective status,index and mode
	replicaIPStatus := make(map[string]*Value)
	for i, v := range addressIPStrings {
		if v != "nil" {
			replicaIPStatus[v] = &Value{index: i, status: replicaStatusStrings[i], mode: "NA"}
		} else {
			// appending address with index to avoid same key conflict
			replicaIPStatus[v+string(i)] = &Value{index: i, status: replicaStatusStrings[i], mode: "NA"}
		}
	}

	// We get the info of the running replicas from the collection.data.
	// We are appending modes if available in collection.data to replicaIPStatus

	replicaInfo := make(map[int]*ReplicaInfo)
	replicaInfo[0] = &ReplicaInfo{"IP", "ACCESSMODE", "STATUS", "NAME", "NODE"}
	replicaInfo[1] = &ReplicaInfo{"---", "-----------", "-------", "-----", "-----"}
	for key := range collection.Data {
		address = append(address, strings.TrimSuffix(strings.TrimPrefix(collection.Data[key].Address, "tcp://"), v1.ReplicaPort))
		mode = append(mode, collection.Data[key].Mode)
		replicaIPStatus[address[key]].mode = mode[key]

	}

	for k, v := range replicaIPStatus {
		// checking if the first three letters is nil or not if it is nil then the ip is not avaiable
		if k[0:3] != "nil" {
			replicaInfo[v.index+2] = &ReplicaInfo{k, v.mode, v.status, "Error fetching Name", "Error Fetching Node"}
		} else {
			replicaInfo[v.index+2] = &ReplicaInfo{"NA", v.mode, v.status, "Error fetching Name", "Error Fetching Node"}
		}
	}

	err = updateReplicasInfo(replicaInfo)
	if err != nil {
		fmt.Println("Error in getting information from K8s. Please try again")
	}

<<<<<<< 47d72fc44065a7693c2b53a6ddd45a234114afdf
	tmpl = template.New("ReplicaInfo")
	tmpl = template.Must(tmpl.Parse(replicaTemplate))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
=======
	if length < replicaCount {
		for i := length; i < (replicaCount); i++ {
			replicaInfo[i+1] = &ReplicaInfo{"NA", "NA", replicaStatus[i], "NA", "NA"}
		}
	}
	tmpl = template.New("ReplicaInfo")
	tmpl = template.Must(tmpl.Parse(replicaTemplate))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
>>>>>>> Proper alignment with the use of tabwriter
	err = tmpl.Execute(w, replicaInfo)
	if err != nil {
		fmt.Println("Unable to display volume info, found error : ", err)
	}
	w.Flush()

	return nil
}
