### Maya API server with Nomad as its orchestration provider

Maya API service is launched from its binary in a dedicated VM. This service is 
the interface for the storage clients to operate on OpenEBS storage. 

> OpenEBS has the concept of VSM (Volume Storage Machine) to provide persistent
storage. Maya API service provides operations w.r.t VSM as a unit.

Notes:

- This guide shows ways to test Maya API service with Nomad as its orchestration provider
- Use of Maya operator is suggested for production / customer usecases
- Maya operator simplifies most of these manual steps into automated ones
- In addition, Maya operator takes care of appropriate release versions

#### Launch Maya API service from its executable

```bash
# from your laptop
$ git clone https://github.com/openebs/mayaserver.git
$ cd mayaserver

$ vagrant up
$ vagrant ssh

# from within the VM
$ make init
$ make
$ make bin
$ nohup m-apiserver up &>mapiserver.log &

# verify the launch
cat mapiserver.log
```

#### Specs to provision OpenEBS VSM 

```
# these are the env variables that needs to be set with 
# appropriate values
$ cat /etc/profile.d/mapiservice.sh 
export DEFAULT_ORCHESTRATOR_NAME="nomad"
export NOMAD_ADDR="http://172.28.128.3:4646"
```

cat my_jiva_vsm.yaml 

```yaml
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: my-jiva-vsm
```

#### Create a VSM


```
# Run this command where maya api service is running

curl -k -H "Content-Type: application/yaml" \
 -XPOST -d"$(cat my_jiva_vsm.yaml)" \
 http://127.0.0.1:5656/latest/volumes/


# sample output
{
  "metadata": {
    "annotations": {
      "evalstatusdesc": "",
      "evalblockedeval": "e6c84d40-abaa-5a47-8faa-a29b147225d3",
      "evalpriority": "50",
      "evaltype": "service",
      "evaltrigger": "job-register",
      "evaljob": "my-jiva-vsm",
      "evalstatus": "complete"
    },
    "creationTimestamp": null,
    "name": "my-jiva-vsm"
  },
  "spec": {
    "AccessModes": null,
    "Capacity": null,
    "ClaimRef": null,
    "OpenEBS": {
      "volumeID": ""
    },
    "PersistentVolumeReclaimPolicy": "",
    "StorageClassName": ""
  },
  "status": {
    "Message": "",
    "Phase": "",
    "Reason": "complete"
  }
}

# sample nomad job spec that gets created

$ nomad inspect m
{
    "Job": {
        "AllAtOnce": false,
        "Constraints": [
            {
                "LTarget": "${attr.kernel.name}",
                "Operand": "=",
                "RTarget": "linux"
            }
        ],
        "CreateIndex": 314,
        "Datacenters": [
            "dc1"
        ],
        "ID": "my-jiva-vsm",
        "JobModifyIndex": 314,
        "Meta": {
            "vsm.openebs.io/cluster-ips": "",
            "vsm.openebs.io/replica-count": "2",
            "vsm.openebs.io/iqn": "iqn.2016-09.com.openebs.jiva:my-jiva-vsm",
            "vsm.openebs.io/replica-ips": "172.28.128.21,172.28.128.19",
            "vsm.openebs.io/targetportals": "172.28.128.22:3260",
            "vsm.openebs.io/controller-status": "",
            "vsm.openebs.io/volume-size": "1G",
            "vsm.openebs.io/controller-ips": "172.28.128.22",
            "vsm.openebs.io/replica-status": ""
        },
        "ModifyIndex": 315,
        "Name": "my-jiva-vsm",
        "ParameterizedJob": null,
        "ParentID": "",
        "Payload": null,
        "Periodic": null,
        "Priority": 50,
        "Region": "global",
        "Status": "pending",
        "StatusDescription": "",
        "TaskGroups": [
            {
                "Constraints": null,
                "Count": 1,
                "EphemeralDisk": {
                    "Migrate": false,
                    "SizeMB": 300,
                    "Sticky": false
                },
                "Meta": null,
                "Name": "fe-jiva-pod",
                "RestartPolicy": {
                    "Attempts": 3,
                    "Delay": 25000000000,
                    "Interval": 300000000000,
                    "Mode": "delay"
                },
                "Tasks": [
                    {
                        "Artifacts": [
                            {
                                "GetterOptions": null,
                                "GetterSource": "https://raw.githubusercontent.com/openebs/jiva/master/scripts/launch-jiva-ctl-with-ip",
                                "RelativeDest": "local/"
                            }
                        ],
                        "Config": {
                            "command": "launch-jiva-ctl-with-ip"
                        },
                        "Constraints": null,
                        "DispatchPayload": null,
                        "Driver": "raw_exec",
                        "Env": {
                            "JIVA_CTL_SUBNET": "24",
                            "JIVA_CTL_VERSION": "openebs/jiva:0.3-RC2",
                            "JIVA_CTL_VOLNAME": "my-jiva-vsm",
                            "JIVA_CTL_VOLSIZE": "1G",
                            "JIVA_CTL_IFACE": "enp0s8",
                            "JIVA_CTL_IP": "172.28.128.22",
                            "JIVA_CTL_NAME": "my-jiva-vsm-fe${NOMAD_ALLOC_INDEX}"
                        },
                        "KillTimeout": 5000000000,
                        "Leader": false,
                        "LogConfig": {
                            "MaxFileSizeMB": 1,
                            "MaxFiles": 3
                        },
                        "Meta": null,
                        "Name": "fe",
                        "Resources": {
                            "CPU": 50,
                            "DiskMB": 0,
                            "IOPS": 0,
                            "MemoryMB": 50,
                            "Networks": [
                                {
                                    "CIDR": "",
                                    "DynamicPorts": null,
                                    "IP": "",
                                    "MBits": 50,
                                    "Public": false,
                                    "ReservedPorts": null
                                }
                            ]
                        },
                        "Services": [],
                        "Templates": [],
                        "User": "",
                        "Vault": null
                    }
                ]
            },
            {
                "Constraints": [
                    {
                        "LTarget": "",
                        "Operand": "distinct_hosts",
                        "RTarget": "true"
                    }
                ],
                "Count": 2,
                "EphemeralDisk": {
                    "Migrate": false,
                    "SizeMB": 300,
                    "Sticky": false
                },
                "Meta": null,
                "Name": "be-jiva-pod",
                "RestartPolicy": {
                    "Attempts": 3,
                    "Delay": 25000000000,
                    "Interval": 300000000000,
                    "Mode": "delay"
                },
                "Tasks": [
                    {
                        "Artifacts": [
                            {
                                "GetterOptions": null,
                                "GetterSource": "https://raw.githubusercontent.com/openebs/jiva/master/scripts/launch-jiva-rep-with-ip",
                                "RelativeDest": "local/"
                            }
                        ],
                        "Config": {
                            "command": "launch-jiva-rep-with-ip"
                        },
                        "Constraints": null,
                        "DispatchPayload": null,
                        "Driver": "raw_exec",
                        "Env": {
                            "JIVA_REP_NETWORK": "host",
                            "JIVA_REP_NAME": "my-jiva-vsm-be${NOMAD_ALLOC_INDEX}",
                            "NOMAD_ALLOC_INDEX": "${NOMAD_ALLOC_INDEX}",
                            "JIVA_REP_VOLSTORE": "/var/openebsmy-jiva-vsm/be${NOMAD_ALLOC_INDEX}",
                            "JIVA_CTL_IP": "172.28.128.22",
                            "JIVA_REP_IFACE": "enp0s8",
                            "JIVA_REP_SUBNET": "24",
                            "JIVA_REP_VOLSIZE": "1G",
                            "JIVA_REP_VERSION": "openebs/jiva:0.3-RC2",
                            "JIVA_REP_IP_1": "172.28.128.19",
                            "JIVA_REP_IP_0": "172.28.128.21",
                            "JIVA_REP_VOLNAME": "my-jiva-vsm"
                        },
                        "KillTimeout": 5000000000,
                        "Leader": false,
                        "LogConfig": {
                            "MaxFileSizeMB": 1,
                            "MaxFiles": 3
                        },
                        "Meta": null,
                        "Name": "be",
                        "Resources": {
                            "CPU": 50,
                            "DiskMB": 0,
                            "IOPS": 0,
                            "MemoryMB": 50,
                            "Networks": [
                                {
                                    "CIDR": "",
                                    "DynamicPorts": null,
                                    "IP": "",
                                    "MBits": 50,
                                    "Public": false,
                                    "ReservedPorts": null
                                }
                            ]
                        },
                        "Services": [],
                        "Templates": [],
                        "User": "",
                        "Vault": null
                    }
                ]
            }
        ],
        "Type": "service",
        "Update": {
            "MaxParallel": 0,
            "Stagger": 0
        },
        "VaultToken": ""
    }
}
```

#### Read a OpenEBS VSM

```
# run this command where maya api service is running

curl http://127.0.0.1:5656/latest/volumes/info/my-jiva-vsm

# sample output
{
  "metadata": {
    "annotations": {
      "vsm.openebs.io\/controller-status": "pending",
      "vsm.openebs.io\/iqn": "iqn.2016-09.com.openebs.jiva:my-jiva-vsm",
      "vsm.openebs.io\/replica-status": "pending",
      "vsm.openebs.io\/replica-ips": "172.28.128.21,172.28.128.19",
      "vsm.openebs.io\/targetportals": "172.28.128.22:3260",
      "vsm.openebs.io\/replica-count": "2",
      "vsm.openebs.io\/controller-ips": "172.28.128.22",
      "vsm.openebs.io\/volume-size": "1G",
      "vsm.openebs.io\/cluster-ips": ""
    },
    "creationTimestamp": null,
    "name": "my-jiva-vsm"
  },
  "spec": {
    "AccessModes": null,
    "Capacity": null,
    "ClaimRef": null,
    "OpenEBS": {
      "volumeID": ""
    },
    "PersistentVolumeReclaimPolicy": "",
    "StorageClassName": ""
  },
  "status": {
    "Message": "",
    "Phase": "",
    "Reason": "pending"
  }
}
```

#### Delete a VSM

```
curl http://127.0.0.1:5656/latest/volumes/delete/my-jiva-vsm
"VSM 'my-jiva-vsm' deleted successfully"

curl http://127.0.0.1:5656/latest/volumes/info/my-jiva-vsm
Unexpected response code: 404 (job not found)
```

#### List VSMs

```
$ curl http://127.0.0.1:5656/latest/volumes/
ListStorage is not implemented by 'orchprovider.mapi.openebs.io/name: nomad'
```
