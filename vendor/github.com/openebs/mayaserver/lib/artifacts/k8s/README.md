## RUN m-apiserver as K8s Pod
```
kubectl create -f maya-apiserver.yaml 
```

## Check the status of the maya-apiserver pod
```
kubectl get pods
kubectl describe pod maya-apiserver
```

(Note the IP address assigned to the pod from the output of the describe command)

## Query the API of the maya-apiserver
```
curl http://10.44.0.1:5656/latest/meta-data/instance-id
```
