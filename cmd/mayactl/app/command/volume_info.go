package command

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	client "github.com/openebs/maya/pkg/client/jiva"
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
================= Replica Details =================
IP            AccessMode               Status
{{range $key, $value := .}}
{{$value.IP}}     {{$value.AccessMode}}                       {{$value.Status}}
{{end}}
===================================================
`
		portalTemplate = `
================= Portal Details ==================
IQN     :   {{.IQN}}
Volume  :   {{.VolumeName}}
Portal  :   {{.Portal}}
Size    :   {{.Size}}
Status  :   {{.Status}}
===================================================
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
	length := len(collection.Data)
	// This case will occur only if user has manually specified zero replica.
	if replicaCount == 0 {
		fmt.Println("None of the replicas are running, please check the volume pod's status by running [kubectl describe pod -l=openebs/replica --all-namespaces] or try again later.")
		return nil
	}
	// We get the info of the running replicas from the collection.data.
	// If there are no replicas running they are either in CrashedLoopBackOff
	// or in Pending or in ImagePullBackoff.In such cases it will show Waiting
	// NA,  NA in Status, access mode and IP fields respectively.
	replicaInfo := make(map[int]*ReplicaInfo)
	for key, _ := range collection.Data {
		address = append(address, strings.TrimSuffix(strings.TrimPrefix(collection.Data[key].Address, "tcp://"), v1.ReplicaPort))
		mode = append(mode, collection.Data[key].Mode)
		replicaInfo[key] = &ReplicaInfo{address[key], mode[key], "Running"}
	}
	if length < replicaCount {
		for i := length; i < (replicaCount); i++ {
			replicaInfo[i] = &ReplicaInfo{"NA", "       NA", "Waiting"}
		}
	}
	tmpl = template.New("ReplicaInfo")
	tmpl = template.Must(tmpl.Parse(replicaTemplate))
	err = tmpl.Execute(os.Stdout, replicaInfo)
	if err != nil {
		fmt.Println("Unable to display volume info, found error : ", err)
	}
	return nil
}
