package nomad

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/helper"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/openebs/maya/types/v1"
)

// Get the job name from a persistent volume claim
func PvcToJobName(pvc *v1.Volume) (string, error) {

	if pvc == nil {
		return "", fmt.Errorf("Nil persistent volume claim provided")
	}

	if pvc.Name == "" {
		return "", fmt.Errorf("Missing VSM name in pvc")
	}

	return pvc.Name, nil
}

// Transform a PersistentVolumeClaim type to Nomad job type
func PvcToJob(pvc *v1.Volume) (*api.Job, error) {

	if pvc == nil {
		return nil, fmt.Errorf("Nil persistent volume claim provided")
	}

	if pvc.Name == "" {
		return nil, fmt.Errorf("Missing VSM name in pvc")
	}

	jivaFEVolSize := v1.GetPVPStorageSize(pvc.Labels)
	jivaBEVolSize := jivaFEVolSize

	// TODO
	// ID is same as Name currently
	// Do we need to think on it ?
	jobName := helper.StringToPtr(pvc.Name)
	region := helper.StringToPtr(v1.GetOrchestratorRegion(pvc.Labels))
	dc := v1.GetOrchestratorDC(pvc.Labels)

	jivaGroupName := "jiva-pod"
	jivaVolName := pvc.Name

	// Set storage size
	feTaskGroup := "fe" + "-" + jivaGroupName
	beTaskGroup := "be" + "-" + jivaGroupName

	// Default storage policy would required 1 FE & 2 BE
	feTaskName := "fe"
	beTaskName := "be"

	jivaFeVersion := v1.GetControllerImage(pvc.Labels)
	jivaNetworkType := v1.GetOrchestratorNetworkType(pvc.Labels)

	jivaBEPersistentStor := v1.GetPVPPersistentPathOnly(pvc.Labels)

	iJivaBECount, err := v1.GetPVPReplicaCountInt(pvc.Labels)
	if err != nil {
		return nil, err
	}

	jivaFeIPs, jivaBeIPs, err := v1.GetPVPVSMIPs(pvc.Labels)
	if err != nil {
		return nil, err
	}

	jivaFeIPArr := strings.Split(jivaFeIPs, ",")
	jivaBeIPArr := strings.Split(jivaBeIPs, ",")
	jivaFeSubnet, err := v1.GetOrchestratorNetworkSubnet(pvc.Labels)
	if err != nil {
		return nil, err
	}

	jivaFeInterface := v1.GetOrchestratorNetworkInterface(pvc.Labels)

	// Meta information will be used to:
	//    1. Persist metadata w.r.t this job
	//
	// NOTE:
	//    This enables to query various info w.r.t job much later.
	// In addition, job's ENV property can source these metadata by interpreting
	// them.
	jobMeta := map[string]string{
		string(v1.ReplicaStatusAPILbl):    "",
		string(v1.ControllerStatusAPILbl): "",
		string(v1.TargetPortalsAPILbl):    jivaFeIPArr[0] + ":" + string(v1.JivaISCSIPortDef),
		string(v1.ClusterIPsAPILbl):       "",
		string(v1.ReplicaIPsAPILbl):       jivaBeIPs,
		string(v1.ControllerIPsAPILbl):    jivaFeIPs,
		string(v1.IQNAPILbl):              string(v1.JivaIqnFormatPrefix) + ":" + jivaVolName,
		string(v1.VolumeSizeAPILbl):       jivaBEVolSize,
		string(v1.ReplicaCountAPILbl):     strconv.Itoa(iJivaBECount),
	}

	// Jiva FE's ENV among other things interpolates Nomad's built-in properties
	feEnv := map[string]string{
		"JIVA_CTL_NAME":    pvc.Name + "-" + feTaskName + "${NOMAD_ALLOC_INDEX}",
		"JIVA_CTL_VERSION": jivaFeVersion,
		"JIVA_CTL_VOLNAME": jivaVolName,
		"JIVA_CTL_VOLSIZE": jivaFEVolSize,
		"JIVA_CTL_IP":      jivaFeIPArr[0],
		"JIVA_CTL_SUBNET":  jivaFeSubnet,
		"JIVA_CTL_IFACE":   jivaFeInterface,
	}

	// Jiva BE's ENV among other things interpolates Nomad's built-in properties
	beEnv := map[string]string{
		"NOMAD_ALLOC_INDEX": "${NOMAD_ALLOC_INDEX}",
		"JIVA_REP_NAME":     pvc.Name + "-" + beTaskName + "${NOMAD_ALLOC_INDEX}",
		"JIVA_CTL_IP":       jivaFeIPArr[0],
		"JIVA_REP_VOLNAME":  jivaVolName,
		"JIVA_REP_VOLSIZE":  jivaBEVolSize,
		"JIVA_REP_VOLSTORE": jivaBEPersistentStor + pvc.Name + "/" + beTaskName + "${NOMAD_ALLOC_INDEX}",
		"JIVA_REP_VERSION":  jivaFeVersion,
		"JIVA_REP_NETWORK":  jivaNetworkType,
		"JIVA_REP_IFACE":    jivaFeInterface,
		"JIVA_REP_SUBNET":   jivaFeSubnet,
	}

	// This sets below variables with backend IP addresses:
	//
	//  1. job's backend's ENV pairs
	//  2. job's META pairs
	err = setBEIPs(beEnv, jobMeta, jivaBeIPArr, iJivaBECount)
	if err != nil {
		return nil, err
	}

	// TODO
	// Transformation from pvc or pv to nomad types & vice-versa:
	//
	//  1. Need an Interface or functional callback defined at
	// lib/api/v1/nomad/ &
	//  2. implemented by the volume plugins that want
	// to be orchestrated by Nomad
	//  3. This transformer instance needs to be injected from
	// volume plugin to orchestrator, in a generic way.

	// Hardcoded logic all the way
	// Nomad specific defaults, hardcoding is OK.
	// However, volume plugin specific stuff is BAD
	return &api.Job{
		Region:      region,
		Name:        jobName,
		ID:          jobName,
		Datacenters: []string{dc},
		Type:        helper.StringToPtr(api.JobTypeService),
		Priority:    helper.IntToPtr(50),
		Constraints: []*api.Constraint{
			api.NewConstraint("${attr.kernel.name}", "=", "linux"),
		},
		Meta: jobMeta,
		TaskGroups: []*api.TaskGroup{
			// jiva frontend
			&api.TaskGroup{
				Name:  helper.StringToPtr(feTaskGroup),
				Count: helper.IntToPtr(1),
				RestartPolicy: &api.RestartPolicy{
					Attempts: helper.IntToPtr(3),
					Interval: helper.TimeToPtr(5 * time.Minute),
					Delay:    helper.TimeToPtr(25 * time.Second),
					Mode:     helper.StringToPtr("delay"),
				},
				Tasks: []*api.Task{
					&api.Task{
						Name:   feTaskName,
						Driver: "raw_exec",
						Resources: &api.Resources{
							CPU:      helper.IntToPtr(50),
							MemoryMB: helper.IntToPtr(50),
							Networks: []*api.NetworkResource{
								&api.NetworkResource{
									MBits: helper.IntToPtr(50),
								},
							},
						},
						Env: feEnv,
						Artifacts: []*api.TaskArtifact{
							&api.TaskArtifact{
								GetterSource: helper.StringToPtr("https://raw.githubusercontent.com/openebs/jiva/master/scripts/launch-jiva-ctl-with-ip"),
								RelativeDest: helper.StringToPtr("local/"),
							},
						},
						Config: map[string]interface{}{
							"command": "launch-jiva-ctl-with-ip",
						},
						LogConfig: &api.LogConfig{
							MaxFiles:      helper.IntToPtr(3),
							MaxFileSizeMB: helper.IntToPtr(1),
						},
					},
				},
			},
			// jiva replica group
			&api.TaskGroup{
				Name: helper.StringToPtr(beTaskGroup),
				// Replica count
				Count: helper.IntToPtr(iJivaBECount),
				// We want the replicas to spread across hosts
				// This ensures high availability
				Constraints: []*api.Constraint{
					api.NewConstraint("", "distinct_hosts", "true"),
				},
				RestartPolicy: &api.RestartPolicy{
					Attempts: helper.IntToPtr(3),
					Interval: helper.TimeToPtr(5 * time.Minute),
					Delay:    helper.TimeToPtr(25 * time.Second),
					Mode:     helper.StringToPtr("delay"),
				},
				// This has multiple replicas as tasks
				Tasks: []*api.Task{
					&api.Task{
						Name:   beTaskName,
						Driver: "raw_exec",
						Resources: &api.Resources{
							CPU:      helper.IntToPtr(50),
							MemoryMB: helper.IntToPtr(50),
							Networks: []*api.NetworkResource{
								&api.NetworkResource{
									MBits: helper.IntToPtr(50),
								},
							},
						},
						Env: beEnv,
						Artifacts: []*api.TaskArtifact{
							&api.TaskArtifact{
								GetterSource: helper.StringToPtr("https://raw.githubusercontent.com/openebs/jiva/master/scripts/launch-jiva-rep-with-ip"),
								RelativeDest: helper.StringToPtr("local/"),
							},
						},
						Config: map[string]interface{}{
							"command": "launch-jiva-rep-with-ip",
						},
						LogConfig: &api.LogConfig{
							MaxFiles:      helper.IntToPtr(3),
							MaxFileSizeMB: helper.IntToPtr(1),
						},
					},
				},
			},
		},
	}, nil
}

// setBEIPs sets jiva backend environment with all backend IP addresses
func setBEIPs(beEnv, jobMeta map[string]string, jivaBeIPArr []string, iJivaBECount int) error {

	if iJivaBECount <= 0 {
		return fmt.Errorf("Invalid VSM Replica count '%d' provided", iJivaBECount)
	}

	if len(jivaBeIPArr) != iJivaBECount {
		return fmt.Errorf("Replica IP count '%d' does not match replica count '%d'", len(jivaBeIPArr), iJivaBECount)
	}

	var k, v string

	for i := 0; i < iJivaBECount; i++ {
		k = string(v1.JivaBackEndIPPrefixLbl) + strconv.Itoa(i)
		v = jivaBeIPArr[i]
		beEnv[k] = v
	}

	return nil
}

// Transform the evaluation of a job to a PersistentVolume
func JobEvalToPv(jobName string, eval *api.Evaluation) (*v1.Volume, error) {

	if eval == nil {
		return nil, fmt.Errorf("Nil job evaluation provided")
	}

	pv := &v1.Volume{}
	pv.Name = jobName

	evalProps := map[string]string{
		"evalpriority":    strconv.Itoa(eval.Priority),
		"evaltype":        eval.Type,
		"evaltrigger":     eval.TriggeredBy,
		"evaljob":         eval.JobID,
		"evalstatus":      eval.Status,
		"evalstatusdesc":  eval.StatusDescription,
		"evalblockedeval": eval.BlockedEval,
	}
	pv.Annotations = evalProps

	pvs := v1.VolumeStatus{
		Message: eval.StatusDescription,
		Reason:  eval.Status,
	}
	pv.Status = pvs

	return pv, nil
}

func MakeJob(name string) (*api.Job, error) {
	if name == "" {
		return nil, fmt.Errorf("Job name required to create a Job")
	}

	return &api.Job{
		Name: helper.StringToPtr(name),
		// TODO
		// ID is same as Name currently
		ID: helper.StringToPtr(name),
	}, nil
}

// Transform a Nomad Job to a PersistentVolume
func JobToPv(job *api.Job) (*v1.Volume, error) {
	if job == nil {
		return nil, fmt.Errorf("Nil job provided")
	}

	pv := &v1.Volume{}
	pv.Name = *job.Name

	pvs := v1.VolumeStatus{
		Message: *job.StatusDescription,
		Reason:  *job.Status,
	}
	pv.Status = pvs

	// Remember we use the job's metadata to persist metadata w.r.t the job
	pv.Annotations = job.Meta

	if *job.Status == structs.JobStatusRunning {
		// Override the status properties only
		pv.Annotations[string(v1.ReplicaStatusAPILbl)] = structs.JobStatusRunning
		pv.Annotations[string(v1.ControllerStatusAPILbl)] = structs.JobStatusRunning
	} else {
		// Override the status properties only
		// TODO
		//    Need to iterate the job taskgroup & set appropriate status rather than
		// a generic status
		pv.Annotations[string(v1.ReplicaStatusAPILbl)] = *job.Status
		pv.Annotations[string(v1.ControllerStatusAPILbl)] = *job.Status
	}

	return pv, nil
}
