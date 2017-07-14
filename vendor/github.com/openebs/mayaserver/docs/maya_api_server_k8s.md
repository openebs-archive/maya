### Maya API server with Kubernetes as its orchestration provider

Maya API server is launched as a Deployment unit in Kubernetes. This service is 
the interface for the storage clients to operate on OpenEBS storage. Typically, 
volume plugins (e.g. K8s Flex Volume driver) act as http clients to Maya API 
service. 

> OpenEBS has the concept of VSM (Volume Storage Machine) to provide persistent
storage. Maya API service provides operations w.r.t VSM as a unit.

**Notes:**

- The specs in this doc point to `test` image(s)
- Use of openebs operator is suggested for production / customer usecases

#### Operator Specs

##### Operator specs for launching Maya API server & security settings

```bash
$ cat maya-api-service-operator.yaml
```

```yaml
# Define the Service Account
# Define the RBAC rules for the Service Account
# Launch the maya-apiserver ( deployment )
# Launch the maya-storagemanager ( deameon set )

# Create Maya Service Account 
apiVersion: v1
kind: ServiceAccount
metadata:
  name: openebs-maya-operator
  namespace: default
---
# Define Role that allows operations on K8s pods/deployments
#  in "default" namespace
# TODO : change to new namespace, for isolated data network
# TODO : the rules should be updated with required group/resources/verb
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  namespace: default
  name: openebs-maya-operator
rules:
- apiGroups: ["*"]
  resources: ["services","pods","deployments", "events"]
  verbs: ["*"]
- apiGroups: ["*"]
  resources: ["persistentvolumes","persistentvolumeclaims"]
  verbs: ["*"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["*"]
---
# Bind the Service Account with the Role Privileges.
# TODO: Check if default account also needs to be there
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: openebs-maya-operator
  namespace: default
subjects:
- kind: ServiceAccount
  name: openebs-maya-operator
  namespace: default
- kind: User
  name: system:serviceaccount:default:default
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: openebs-maya-operator
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: maya-apiserver
  namespace: default
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: maya-apiserver
    spec:
      serviceAccountName: openebs-maya-operator
      containers:
      - name: maya-apiserver
        imagePullPolicy: Always
        image: openebs/m-apiserver:test
        ports:
        - containerPort: 5656
---
apiVersion: v1
kind: Service
metadata:
  name: maya-apiserver-service
spec:
  ports:
  - name: api
    port: 5656
    protocol: TCP
    targetPort: 5656
  selector:
    name: maya-apiserver
  sessionAffinity: None
```

##### Use kubectl to launch the operator

```bash
kubectl create -f maya-api-service-operator.yaml
```

- Get the IP address of maya-apiserver Pod

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get pod
NAME                                   READY     STATUS    RESTARTS   AGE
maya-apiserver-2275666786-h0qqw        1/1       Running   0          53m

ubuntu@kubemaster-01:/vagrant$ kubectl get pod/maya-apiserver-2275666786-h0qqw -o json
```

```json
{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "annotations": {
            "kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\",\"namespace\":\"default\",\"name\":\"maya-apiserver-2275666786\",\"uid\":\"4479670f-66df-11e7-833f-021c6f7dbe9d\",\"apiVersion\":\"extensions\",\"resourceVersion\":\"8605\"}}\n"
        },
        "creationTimestamp": "2017-07-12T08:51:17Z",
        "generateName": "maya-apiserver-2275666786-",
        "labels": {
            "name": "maya-apiserver",
            "pod-template-hash": "2275666786"
        },
        "name": "maya-apiserver-2275666786-h0qqw",
        "namespace": "default",
        "ownerReferences": [
            {
                "apiVersion": "extensions/v1beta1",
                "blockOwnerDeletion": true,
                "controller": true,
                "kind": "ReplicaSet",
                "name": "maya-apiserver-2275666786",
                "uid": "4479670f-66df-11e7-833f-021c6f7dbe9d"
            }
        ],
        "resourceVersion": "8819",
        "selfLink": "/api/v1/namespaces/default/pods/maya-apiserver-2275666786-h0qqw",
        "uid": "447b7161-66df-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "containers": [
            {
                "image": "openebs/m-apiserver:test",
                "imagePullPolicy": "Always",
                "name": "maya-apiserver",
                "ports": [
                    {
                        "containerPort": 5656,
                        "protocol": "TCP"
                    }
                ],
                "resources": {},
                "terminationMessagePath": "/dev/termination-log",
                "terminationMessagePolicy": "File",
                "volumeMounts": [
                    {
                        "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                        "name": "openebs-maya-operator-token-ww55v",
                        "readOnly": true
                    }
                ]
            }
        ],
        "dnsPolicy": "ClusterFirst",
        "nodeName": "kubeminion-01",
        "restartPolicy": "Always",
        "schedulerName": "default-scheduler",
        "securityContext": {},
        "serviceAccount": "openebs-maya-operator",
        "serviceAccountName": "openebs-maya-operator",
        "terminationGracePeriodSeconds": 30,
        "tolerations": [
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/notReady",
                "operator": "Exists",
                "tolerationSeconds": 300
            },
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/unreachable",
                "operator": "Exists",
                "tolerationSeconds": 300
            }
        ],
        "volumes": [
            {
                "name": "openebs-maya-operator-token-ww55v",
                "secret": {
                    "defaultMode": 420,
                    "secretName": "openebs-maya-operator-token-ww55v"
                }
            }
        ]
    },
    "status": {
        "conditions": [
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T08:51:19Z",
                "status": "True",
                "type": "Initialized"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T08:53:29Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T08:51:17Z",
                "status": "True",
                "type": "PodScheduled"
            }
        ],
        "containerStatuses": [
            {
                "containerID": "docker://d3d88860d0c4cd0dc05414fb791b173a7877146bf7d41d40c201fa1dfce0c74a",
                "image": "openebs/m-apiserver:test",
                "imageID": "docker://sha256:214b9a8ae8d166ff9b982ac5bae83bf828f62efebaa8204423b6cc133310487e",
                "lastState": {},
                "name": "maya-apiserver",
                "ready": true,
                "restartCount": 0,
                "state": {
                    "running": {
                        "startedAt": "2017-07-12T08:53:27Z"
                    }
                }
            }
        ],
        "hostIP": "172.28.128.10",
        "phase": "Running",
        "podIP": "10.44.0.1",
        "qosClass": "BestEffort",
        "startTime": "2017-07-12T08:51:19Z"
    }
}
```

##### Create yaml specs to launch VSM as K8s deployments & K8s service

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: my-jiva-vsm
```

- Alternatively, a sample specs with specific volume size & single replica

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: my-jiva-vsm
  labels:
    volumeprovisioner.mapi.openebs.io/storage-size: 2G
    volumeprovisioner.mapi.openebs.io/replica-count: 1
```

```bash
curl -k -H "Content-Type: application/yaml" \
  -XPOST -d"$(cat my-jiva-vsm.yaml)" \
  http://10.44.0.1:5656/latest/volumes/
```

- One gets the VSM `name` echoed back !!

```json
{
  "metadata": {
    "creationTimestamp": null,
    "name": "my-2-jiva-vsm"
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
    "Reason": ""
  }
}
```

##### Read an existing VSM

```bash
curl http://10.44.0.1:5656/latest/volumes/info/<vsm-name>

# e.g.

curl http://10.44.0.1:5656/latest/volumes/info/my-2-jiva-vsm
```

```json
{
  "metadata": {
    "annotations": {
      "vsm.openebs.io\/controller-ips": "10.44.0.2",
      "vsm.openebs.io\/cluster-ips": "10.111.100.12",
      "vsm.openebs.io\/iqn": "iqn.2016-09.com.openebs.jiva:my-2-jiva-vsm",
      "vsm.openebs.io\/replica-count": "2",
      "vsm.openebs.io\/volume-size": "1G",
      "vsm.openebs.io\/controller-status": "Running",
      "vsm.openebs.io\/replica-ips": "10.44.0.3,10.36.0.2",
      "vsm.openebs.io\/replica-status": "Running,Running",
      "vsm.openebs.io\/targetportals": "10.111.100.12:3260"
    },
    "creationTimestamp": null,
    "name": "my-2-jiva-vsm"
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
    "Reason": ""
  }
}
```

##### Delete an existing VSM

```bash
curl http://10.44.0.1:5656/latest/volumes/delete/<vsm-name>

# e.g.

curl http://10.44.0.1:5656/latest/volumes/delete/my-2-jiva-vsm
```

```
"VSM 'my-2-jiva-vsm' deleted successfully"
```

##### List all VSMs

```bash
curl http://10.44.0.1:5656/latest/volumes/
```

```json
{
  "items": [
    {
      "metadata": {
        "annotations": {
          "vsm.openebs.io\/volume-size": "1G",
          "vsm.openebs.io\/controller-ips": "10.44.0.2",
          "vsm.openebs.io\/controller-status": "Running",
          "vsm.openebs.io\/cluster-ips": "10.111.100.12",
          "vsm.openebs.io\/iqn": "iqn.2016-09.com.openebs.jiva:my-2-jiva-vsm",
          "vsm.openebs.io\/replica-count": "2",
          "vsm.openebs.io\/replica-status": "Running,Running",
          "vsm.openebs.io\/targetportals": "10.111.100.12:3260",
          "vsm.openebs.io\/replica-ips": "10.44.0.3,10.36.0.2"
        },
        "creationTimestamp": null,
        "name": "my-2-jiva-vsm"
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
        "Reason": ""
      }
    }
  ],
  "metadata": {
    
  }
}
```

##### Verify the Service

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get service
NAME                     CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
kubernetes               10.96.0.1       <none>        443/TCP             2h
maya-apiserver-service   10.97.233.57    <none>        5656/TCP            32m
my-2-jiva-vsm-ctrl-svc   10.111.100.12   <none>        3260/TCP,9501/TCP   9m
```

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get service/my-2-jiva-vsm-ctrl-svc -o json
```

```json
{
    "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
        "creationTimestamp": "2017-07-12T09:14:51Z",
        "labels": {
            "openebs/controller-service": "jiva-controller-service",
            "openebs/volume-provisioner": "jiva",
            "vsm": "my-2-jiva-vsm"
        },
        "name": "my-2-jiva-vsm-ctrl-svc",
        "namespace": "default",
        "resourceVersion": "10487",
        "selfLink": "/api/v1/namespaces/default/services/my-2-jiva-vsm-ctrl-svc",
        "uid": "8f4325fd-66e2-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "clusterIP": "10.111.100.12",
        "ports": [
            {
                "name": "iscsi",
                "port": 3260,
                "protocol": "TCP",
                "targetPort": 3260
            },
            {
                "name": "api",
                "port": 9501,
                "protocol": "TCP",
                "targetPort": 9501
            }
        ],
        "selector": {
            "openebs/controller": "jiva-controller",
            "vsm": "my-2-jiva-vsm"
        },
        "sessionAffinity": "None",
        "type": "ClusterIP"
    },
    "status": {
        "loadBalancer": {}
    }
}
```

##### Verify the Deployments

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get deploy
NAME                  DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
my-2-jiva-vsm-ctrl    1         1         1            1           5m
my-2-jiva-vsm-rep     2         2         2            2           5m
```

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get deploy/my-2-jiva-vsm-ctrl -o json
```

```json
{
    "apiVersion": "extensions/v1beta1",
    "kind": "Deployment",
    "metadata": {
        "annotations": {
            "deployment.kubernetes.io/revision": "1"
        },
        "creationTimestamp": "2017-07-12T09:14:51Z",
        "generation": 1,
        "labels": {
            "openebs/controller": "jiva-controller",
            "openebs/volume-provisioner": "jiva",
            "vsm": "my-2-jiva-vsm"
        },
        "name": "my-2-jiva-vsm-ctrl",
        "namespace": "default",
        "resourceVersion": "10594",
        "selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/my-2-jiva-vsm-ctrl",
        "uid": "8f498a0e-66e2-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "replicas": 1,
        "selector": {
            "matchLabels": {
                "openebs/controller": "jiva-controller",
                "vsm": "my-2-jiva-vsm"
            }
        },
        "strategy": {
            "rollingUpdate": {
                "maxSurge": 1,
                "maxUnavailable": 1
            },
            "type": "RollingUpdate"
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "openebs/controller": "jiva-controller",
                    "vsm": "my-2-jiva-vsm"
                }
            },
            "spec": {
                "containers": [
                    {
                        "args": [
                            "controller",
                            "--frontend",
                            "gotgt",
                            "--clusterIP",
                            "10.111.100.12",
                            "my-2-jiva-vsm"
                        ],
                        "command": [
                            "launch"
                        ],
                        "image": "openebs/jiva:0.3-RC2",
                        "imagePullPolicy": "IfNotPresent",
                        "name": "my-2-jiva-vsm-ctrl-con",
                        "ports": [
                            {
                                "containerPort": 3260,
                                "protocol": "TCP"
                            },
                            {
                                "containerPort": 9501,
                                "protocol": "TCP"
                            }
                        ],
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File"
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "restartPolicy": "Always",
                "schedulerName": "default-scheduler",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30
            }
        }
    },
    "status": {
        "availableReplicas": 1,
        "conditions": [
            {
                "lastTransitionTime": "2017-07-12T09:14:51Z",
                "lastUpdateTime": "2017-07-12T09:14:51Z",
                "message": "Deployment has minimum availability.",
                "reason": "MinimumReplicasAvailable",
                "status": "True",
                "type": "Available"
            }
        ],
        "observedGeneration": 1,
        "readyReplicas": 1,
        "replicas": 1,
        "updatedReplicas": 1
    }
}
```

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get deploy/my-2-jiva-vsm-rep -o json
```

```json
{
    "apiVersion": "extensions/v1beta1",
    "kind": "Deployment",
    "metadata": {
        "annotations": {
            "deployment.kubernetes.io/revision": "1"
        },
        "creationTimestamp": "2017-07-12T09:14:51Z",
        "generation": 1,
        "labels": {
            "openebs/replica": "jiva-replica",
            "openebs/volume-provisioner": "jiva",
            "vsm": "my-2-jiva-vsm"
        },
        "name": "my-2-jiva-vsm-rep",
        "namespace": "default",
        "resourceVersion": "10596",
        "selfLink": "/apis/extensions/v1beta1/namespaces/default/deployments/my-2-jiva-vsm-rep",
        "uid": "8f4e0da4-66e2-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "replicas": 2,
        "selector": {
            "matchLabels": {
                "openebs/replica": "jiva-replica",
                "vsm": "my-2-jiva-vsm"
            }
        },
        "strategy": {
            "rollingUpdate": {
                "maxSurge": 1,
                "maxUnavailable": 1
            },
            "type": "RollingUpdate"
        },
        "template": {
            "metadata": {
                "creationTimestamp": null,
                "labels": {
                    "openebs/replica": "jiva-replica",
                    "vsm": "my-2-jiva-vsm"
                }
            },
            "spec": {
                "affinity": {
                    "podAntiAffinity": {
                        "requiredDuringSchedulingIgnoredDuringExecution": [
                            {
                                "labelSelector": {
                                    "matchLabels": {
                                        "openebs/replica": "jiva-replica",
                                        "vsm": "my-2-jiva-vsm"
                                    }
                                },
                                "topologyKey": "kubernetes.io/hostname"
                            }
                        ]
                    }
                },
                "containers": [
                    {
                        "args": [
                            "replica",
                            "--frontendIP",
                            "10.111.100.12",
                            "--size",
                            "1G",
                            "/openebs"
                        ],
                        "command": [
                            "launch"
                        ],
                        "image": "openebs/jiva:0.3-RC2",
                        "imagePullPolicy": "IfNotPresent",
                        "name": "my-2-jiva-vsm-rep-con",
                        "ports": [
                            {
                                "containerPort": 9502,
                                "protocol": "TCP"
                            },
                            {
                                "containerPort": 9503,
                                "protocol": "TCP"
                            },
                            {
                                "containerPort": 9504,
                                "protocol": "TCP"
                            }
                        ],
                        "resources": {},
                        "terminationMessagePath": "/dev/termination-log",
                        "terminationMessagePolicy": "File",
                        "volumeMounts": [
                            {
                                "mountPath": "/openebs",
                                "name": "openebs"
                            }
                        ]
                    }
                ],
                "dnsPolicy": "ClusterFirst",
                "restartPolicy": "Always",
                "schedulerName": "default-scheduler",
                "securityContext": {},
                "terminationGracePeriodSeconds": 30,
                "volumes": [
                    {
                        "hostPath": {
                            "path": "/var/openebs/my-2-jiva-vsm/openebs"
                        },
                        "name": "openebs"
                    }
                ]
            }
        }
    },
    "status": {
        "availableReplicas": 2,
        "conditions": [
            {
                "lastTransitionTime": "2017-07-12T09:15:09Z",
                "lastUpdateTime": "2017-07-12T09:15:09Z",
                "message": "Deployment has minimum availability.",
                "reason": "MinimumReplicasAvailable",
                "status": "True",
                "type": "Available"
            }
        ],
        "observedGeneration": 1,
        "readyReplicas": 2,
        "replicas": 2,
        "updatedReplicas": 2
    }
}
```

##### Verify the Pods

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get pods
NAME                                   READY     STATUS    RESTARTS   AGE
my-2-jiva-vsm-ctrl-343520736-kq87w     1/1       Running   0          11m
my-2-jiva-vsm-rep-1390161198-6qsph     1/1       Running   0          11m
my-2-jiva-vsm-rep-1390161198-cshrg     1/1       Running   0          11m
```

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get pod/my-2-jiva-vsm-ctrl-343520736-kq87w -o json
```

```json
{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "annotations": {
            "kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\",\"namespace\":\"default\",\"name\":\"my-2-jiva-vsm-ctrl-343520736\",\"uid\":\"8f4cb8d1-66e2-11e7-833f-021c6f7dbe9d\",\"apiVersion\":\"extensions\",\"resourceVersion\":\"10490\"}}\n"
        },
        "creationTimestamp": "2017-07-12T09:14:51Z",
        "generateName": "my-2-jiva-vsm-ctrl-343520736-",
        "labels": {
            "openebs/controller": "jiva-controller",
            "pod-template-hash": "343520736",
            "vsm": "my-2-jiva-vsm"
        },
        "name": "my-2-jiva-vsm-ctrl-343520736-kq87w",
        "namespace": "default",
        "ownerReferences": [
            {
                "apiVersion": "extensions/v1beta1",
                "blockOwnerDeletion": true,
                "controller": true,
                "kind": "ReplicaSet",
                "name": "my-2-jiva-vsm-ctrl-343520736",
                "uid": "8f4cb8d1-66e2-11e7-833f-021c6f7dbe9d"
            }
        ],
        "resourceVersion": "10590",
        "selfLink": "/api/v1/namespaces/default/pods/my-2-jiva-vsm-ctrl-343520736-kq87w",
        "uid": "8f500281-66e2-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "containers": [
            {
                "args": [
                    "controller",
                    "--frontend",
                    "gotgt",
                    "--clusterIP",
                    "10.111.100.12",
                    "my-2-jiva-vsm"
                ],
                "command": [
                    "launch"
                ],
                "image": "openebs/jiva:0.3-RC2",
                "imagePullPolicy": "IfNotPresent",
                "name": "my-2-jiva-vsm-ctrl-con",
                "ports": [
                    {
                        "containerPort": 3260,
                        "protocol": "TCP"
                    },
                    {
                        "containerPort": 9501,
                        "protocol": "TCP"
                    }
                ],
                "resources": {},
                "terminationMessagePath": "/dev/termination-log",
                "terminationMessagePolicy": "File",
                "volumeMounts": [
                    {
                        "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                        "name": "default-token-0v1tp",
                        "readOnly": true
                    }
                ]
            }
        ],
        "dnsPolicy": "ClusterFirst",
        "nodeName": "kubeminion-01",
        "restartPolicy": "Always",
        "schedulerName": "default-scheduler",
        "securityContext": {},
        "serviceAccount": "default",
        "serviceAccountName": "default",
        "terminationGracePeriodSeconds": 30,
        "tolerations": [
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/notReady",
                "operator": "Exists",
                "tolerationSeconds": 300
            },
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/unreachable",
                "operator": "Exists",
                "tolerationSeconds": 300
            }
        ],
        "volumes": [
            {
                "name": "default-token-0v1tp",
                "secret": {
                    "defaultMode": 420,
                    "secretName": "default-token-0v1tp"
                }
            }
        ]
    },
    "status": {
        "conditions": [
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:14:51Z",
                "status": "True",
                "type": "Initialized"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:15:31Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:14:51Z",
                "status": "True",
                "type": "PodScheduled"
            }
        ],
        "containerStatuses": [
            {
                "containerID": "docker://35f00ee78c791e180e93a4fbbddb323919bde6b76b82fa61b673185fbe5f943e",
                "image": "openebs/jiva:0.3-RC2",
                "imageID": "docker://sha256:ab153ccaa15e55ee13b294905c8181a946ffab9765fc7267e38841fe94412d8b",
                "lastState": {},
                "name": "my-2-jiva-vsm-ctrl-con",
                "ready": true,
                "restartCount": 0,
                "state": {
                    "running": {
                        "startedAt": "2017-07-12T09:15:31Z"
                    }
                }
            }
        ],
        "hostIP": "172.28.128.10",
        "phase": "Running",
        "podIP": "10.44.0.2",
        "qosClass": "BestEffort",
        "startTime": "2017-07-12T09:14:51Z"
    }
}
```

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get pod/my-2-jiva-vsm-rep-1390161198-6qsph -o json
```

```json
{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "annotations": {
            "kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\",\"namespace\":\"default\",\"name\":\"my-2-jiva-vsm-rep-1390161198\",\"uid\":\"8f50966f-66e2-11e7-833f-021c6f7dbe9d\",\"apiVersion\":\"extensions\",\"resourceVersion\":\"10495\"}}\n"
        },
        "creationTimestamp": "2017-07-12T09:14:51Z",
        "generateName": "my-2-jiva-vsm-rep-1390161198-",
        "labels": {
            "openebs/replica": "jiva-replica",
            "pod-template-hash": "1390161198",
            "vsm": "my-2-jiva-vsm"
        },
        "name": "my-2-jiva-vsm-rep-1390161198-6qsph",
        "namespace": "default",
        "ownerReferences": [
            {
                "apiVersion": "extensions/v1beta1",
                "blockOwnerDeletion": true,
                "controller": true,
                "kind": "ReplicaSet",
                "name": "my-2-jiva-vsm-rep-1390161198",
                "uid": "8f50966f-66e2-11e7-833f-021c6f7dbe9d"
            }
        ],
        "resourceVersion": "10592",
        "selfLink": "/api/v1/namespaces/default/pods/my-2-jiva-vsm-rep-1390161198-6qsph",
        "uid": "8f56785c-66e2-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "affinity": {
            "podAntiAffinity": {
                "requiredDuringSchedulingIgnoredDuringExecution": [
                    {
                        "labelSelector": {
                            "matchLabels": {
                                "openebs/replica": "jiva-replica",
                                "vsm": "my-2-jiva-vsm"
                            }
                        },
                        "topologyKey": "kubernetes.io/hostname"
                    }
                ]
            }
        },
        "containers": [
            {
                "args": [
                    "replica",
                    "--frontendIP",
                    "10.111.100.12",
                    "--size",
                    "1G",
                    "/openebs"
                ],
                "command": [
                    "launch"
                ],
                "image": "openebs/jiva:0.3-RC2",
                "imagePullPolicy": "IfNotPresent",
                "name": "my-2-jiva-vsm-rep-con",
                "ports": [
                    {
                        "containerPort": 9502,
                        "protocol": "TCP"
                    },
                    {
                        "containerPort": 9503,
                        "protocol": "TCP"
                    },
                    {
                        "containerPort": 9504,
                        "protocol": "TCP"
                    }
                ],
                "resources": {},
                "terminationMessagePath": "/dev/termination-log",
                "terminationMessagePolicy": "File",
                "volumeMounts": [
                    {
                        "mountPath": "/openebs",
                        "name": "openebs"
                    },
                    {
                        "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                        "name": "default-token-0v1tp",
                        "readOnly": true
                    }
                ]
            }
        ],
        "dnsPolicy": "ClusterFirst",
        "nodeName": "kubeminion-01",
        "restartPolicy": "Always",
        "schedulerName": "default-scheduler",
        "securityContext": {},
        "serviceAccount": "default",
        "serviceAccountName": "default",
        "terminationGracePeriodSeconds": 30,
        "tolerations": [
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/notReady",
                "operator": "Exists",
                "tolerationSeconds": 300
            },
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/unreachable",
                "operator": "Exists",
                "tolerationSeconds": 300
            }
        ],
        "volumes": [
            {
                "hostPath": {
                    "path": "/var/openebs/my-2-jiva-vsm/openebs"
                },
                "name": "openebs"
            },
            {
                "name": "default-token-0v1tp",
                "secret": {
                    "defaultMode": 420,
                    "secretName": "default-token-0v1tp"
                }
            }
        ]
    },
    "status": {
        "conditions": [
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:14:52Z",
                "status": "True",
                "type": "Initialized"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:15:31Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:14:52Z",
                "status": "True",
                "type": "PodScheduled"
            }
        ],
        "containerStatuses": [
            {
                "containerID": "docker://8163434766e92f5aa3114f3bb1da7b4652d8b83219605b5b068ae17ace4c1c6e",
                "image": "openebs/jiva:0.3-RC2",
                "imageID": "docker://sha256:ab153ccaa15e55ee13b294905c8181a946ffab9765fc7267e38841fe94412d8b",
                "lastState": {},
                "name": "my-2-jiva-vsm-rep-con",
                "ready": true,
                "restartCount": 0,
                "state": {
                    "running": {
                        "startedAt": "2017-07-12T09:15:31Z"
                    }
                }
            }
        ],
        "hostIP": "172.28.128.10",
        "phase": "Running",
        "podIP": "10.44.0.3",
        "qosClass": "BestEffort",
        "startTime": "2017-07-12T09:14:52Z"
    }
}
```

```bash
ubuntu@kubemaster-01:/vagrant$ kubectl get pod/my-2-jiva-vsm-rep-1390161198-cshrg -o json
```

```json
{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "annotations": {
            "kubernetes.io/created-by": "{\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\",\"namespace\":\"default\",\"name\":\"my-2-jiva-vsm-rep-1390161198\",\"uid\":\"8f50966f-66e2-11e7-833f-021c6f7dbe9d\",\"apiVersion\":\"extensions\",\"resourceVersion\":\"10495\"}}\n"
        },
        "creationTimestamp": "2017-07-12T09:14:51Z",
        "generateName": "my-2-jiva-vsm-rep-1390161198-",
        "labels": {
            "openebs/replica": "jiva-replica",
            "pod-template-hash": "1390161198",
            "vsm": "my-2-jiva-vsm"
        },
        "name": "my-2-jiva-vsm-rep-1390161198-cshrg",
        "namespace": "default",
        "ownerReferences": [
            {
                "apiVersion": "extensions/v1beta1",
                "blockOwnerDeletion": true,
                "controller": true,
                "kind": "ReplicaSet",
                "name": "my-2-jiva-vsm-rep-1390161198",
                "uid": "8f50966f-66e2-11e7-833f-021c6f7dbe9d"
            }
        ],
        "resourceVersion": "10554",
        "selfLink": "/api/v1/namespaces/default/pods/my-2-jiva-vsm-rep-1390161198-cshrg",
        "uid": "8f55f23b-66e2-11e7-833f-021c6f7dbe9d"
    },
    "spec": {
        "affinity": {
            "podAntiAffinity": {
                "requiredDuringSchedulingIgnoredDuringExecution": [
                    {
                        "labelSelector": {
                            "matchLabels": {
                                "openebs/replica": "jiva-replica",
                                "vsm": "my-2-jiva-vsm"
                            }
                        },
                        "topologyKey": "kubernetes.io/hostname"
                    }
                ]
            }
        },
        "containers": [
            {
                "args": [
                    "replica",
                    "--frontendIP",
                    "10.111.100.12",
                    "--size",
                    "1G",
                    "/openebs"
                ],
                "command": [
                    "launch"
                ],
                "image": "openebs/jiva:0.3-RC2",
                "imagePullPolicy": "IfNotPresent",
                "name": "my-2-jiva-vsm-rep-con",
                "ports": [
                    {
                        "containerPort": 9502,
                        "protocol": "TCP"
                    },
                    {
                        "containerPort": 9503,
                        "protocol": "TCP"
                    },
                    {
                        "containerPort": 9504,
                        "protocol": "TCP"
                    }
                ],
                "resources": {},
                "terminationMessagePath": "/dev/termination-log",
                "terminationMessagePolicy": "File",
                "volumeMounts": [
                    {
                        "mountPath": "/openebs",
                        "name": "openebs"
                    },
                    {
                        "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                        "name": "default-token-0v1tp",
                        "readOnly": true
                    }
                ]
            }
        ],
        "dnsPolicy": "ClusterFirst",
        "nodeName": "kubeminion-02",
        "restartPolicy": "Always",
        "schedulerName": "default-scheduler",
        "securityContext": {},
        "serviceAccount": "default",
        "serviceAccountName": "default",
        "terminationGracePeriodSeconds": 30,
        "tolerations": [
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/notReady",
                "operator": "Exists",
                "tolerationSeconds": 300
            },
            {
                "effect": "NoExecute",
                "key": "node.alpha.kubernetes.io/unreachable",
                "operator": "Exists",
                "tolerationSeconds": 300
            }
        ],
        "volumes": [
            {
                "hostPath": {
                    "path": "/var/openebs/my-2-jiva-vsm/openebs"
                },
                "name": "openebs"
            },
            {
                "name": "default-token-0v1tp",
                "secret": {
                    "defaultMode": 420,
                    "secretName": "default-token-0v1tp"
                }
            }
        ]
    },
    "status": {
        "conditions": [
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:14:52Z",
                "status": "True",
                "type": "Initialized"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:15:09Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2017-07-12T09:14:52Z",
                "status": "True",
                "type": "PodScheduled"
            }
        ],
        "containerStatuses": [
            {
                "containerID": "docker://ecedfa6e4e511a493e92257c79f19e1fcc4dc0e9be96f24b58186c67256c576c",
                "image": "openebs/jiva:0.3-RC2",
                "imageID": "docker://sha256:ab153ccaa15e55ee13b294905c8181a946ffab9765fc7267e38841fe94412d8b",
                "lastState": {},
                "name": "my-2-jiva-vsm-rep-con",
                "ready": true,
                "restartCount": 0,
                "state": {
                    "running": {
                        "startedAt": "2017-07-12T09:15:07Z"
                    }
                }
            }
        ],
        "hostIP": "172.28.128.11",
        "phase": "Running",
        "podIP": "10.36.0.2",
        "qosClass": "BestEffort",
        "startTime": "2017-07-12T09:14:52Z"
    }
}
```
