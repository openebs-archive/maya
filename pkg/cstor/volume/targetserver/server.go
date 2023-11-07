/*
Copyright 2019 The OpenEBS Authors

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

package targetserver

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	cstorv1alpha1 "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	"github.com/pkg/errors"
)

var (
	endOfData          = "\r\n"
	respOk             = "Ok"
	respErr            = "Err"
	volumeMgmtUnixSock = "/var/run/volume_mgmt_sock"
)

// IsSafeToRetry will retrun true if error type is EINTR, EAGAIN or EWOULDBLOCK
func IsSafeToRetry(err error) bool {
	// For more information https://golang.org/pkg/syscall/#Errno
	if err == syscall.EINTR ||
		err == syscall.EAGAIN ||
		err == syscall.EWOULDBLOCK {
		return true
	}
	return false
}

/* Client will send below data to process
 * Ex JSON data:
 * {"replicaId":"6061","replicaZvolGuid":"6061",
 * "volumeName":"vol1","replicationFactor":3,"consistencyFactor":2}
 * and server will process it and returns Ok in case of not having error else it
 * returns Err
 * Response will be either "Ok" or "Err"
 */

// Reader reads the data from wire untill error or endOfData occurs
// Reader will break only when client is sending \r\n or EOF occured
func Reader(r io.Reader) (string, error) {
	buf := make([]byte, 1024)
	//collect bytes into fulllines buffer till the end of line character is reached
	completeBytes := []byte{}
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", errors.Wrapf(err, "failed to read data on wire")
		}
		if n > 0 {
			completeBytes = append(completeBytes, buf[0:n]...)
			if strings.HasSuffix(string(completeBytes), endOfData) {
				// Since we are ending only when \r\n bytes are present in
				// completeBytes to extract \r\n need to perform below steps
				lines := strings.Split(string(completeBytes), endOfData)
				return lines[0], nil
			}
		}
	}
	// Code will reach here only when EOF happens
	return "", errors.Errorf("failed to read data connection closed")
}

// ServeRequest process the request from the client
func ServeRequest(conn net.Conn, kubeClient *cstorv1alpha1.Kubeclient) {
	var err error
	var readData string
	defer func() {
		if err != nil {
			_, er := conn.Write([]byte(respErr + endOfData))
			if er != nil {
				klog.Errorf("failed to inform to client")
			}
		} else {
			_, er := conn.Write([]byte(respOk + endOfData))
			if er != nil {
				klog.Errorf("failed to inform to client")
			}
		}
		conn.Close()
	}()
	//example readData:
	// {"replicaId":"6061","replicaZvolGuid":"6061","volumeName":"vol1",
	// "replicationFactor":3,"consistencyFactor":2}
	readData, err = Reader(conn)
	if err != nil {
		klog.Errorf("failed to read data: {%v}", err)
		return
	}
	replicationData := &cstorv1alpha1.CVReplicationDetails{}
	err = json.Unmarshal([]byte(readData), replicationData)
	if err != nil {
		klog.Errorf("failed to unmarshal replication data {%v}", err)
		return
	}
	err = replicationData.UpdateCVWithReplicationDetails(kubeClient)
	if err != nil {
		klog.Errorf("failed to update cstorvolume {%s} with details {%v}"+
			" error: {%v}", replicationData.VolumeName,
			replicationData, err)
		return
	}
}

// StartTargetServer starts the UnixDomainServer
func StartTargetServer(kubeConfig string) {

	var namespace string
	for {
		klog.Info("Waiting for namespace to be populated")
		if cstorv1alpha1.TargetNamespace != "" {
			namespace = cstorv1alpha1.TargetNamespace
			break
		}
		// Sleep of 3 secs is good enough since target deployment will be created
		// only when volume is created
		time.Sleep(time.Second * 3)
	}
	klog.Infof("CstorVolume namespace %s", namespace)

	klog.Info("Starting unix domain server")
	if err := os.RemoveAll(string(volumeMgmtUnixSock)); err != nil {
		klog.Fatalf("failed to clear path: {%v}", err)
	}

	listen, err := net.Listen("unix", volumeMgmtUnixSock)
	if err != nil {
		klog.Fatalf("listen error: {%v}", err)
	}
	defer listen.Close()

	// Since we are reading kubeClient there is no need to taking lock
	kubeClient := cstorv1alpha1.NewKubeclient(
		cstorv1alpha1.WithKubeConfigPath(kubeConfig)).
		WithNamespace(namespace)

	for {
		sockFd, err := listen.Accept()
		if IsSafeToRetry(err) {
			klog.Errorf("failed to accept error: {%v} will continue...", err)
			continue
		}
		// If it is unknown error exit the process
		if err != nil {
			klog.Fatalf("failed to accept error: {%v}", err)
		}
		go ServeRequest(sockFd, kubeClient)
	}
}
