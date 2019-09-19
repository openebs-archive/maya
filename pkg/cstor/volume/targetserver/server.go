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
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"k8s.io/klog"

	cstorv1alpha1 "github.com/openebs/maya/pkg/cstor/volume/v1alpha1"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"github.com/pkg/errors"
)

var (
	endOfLine          = "\r\n"
	respOk             = "Ok"
	respErr            = "Err"
	volumeMgmtUnixSock = "/var/run/volume_mgmt_sock"
)

// Reader reads the data from wire untill error or endOfLine occurs
func Reader(r io.Reader) (string, error) {
	req := []string{}
	//collect bytes into fulllines buffer till the end of line character is reached
	completeBytes := []byte{}
	for {
		buf := make([]byte, 1024)
		n, err := r.Read(buf[:])
		if err != nil {
			if err == io.EOF {
				klog.Info("Reached End Of file")
				break
			}
			return "", errors.Wrapf(err, "failed to read data on wire")
		}
		if n > 0 {
			completeBytes = append(completeBytes, buf[0:n]...)
			if strings.HasSuffix(string(completeBytes), endOfLine) {
				lines := strings.Split(string(completeBytes), endOfLine)
				for _, line := range lines {
					if len(line) != 0 {
						req = append(req, line+endOfLine)
					}
				}
				break
			}
		}
	}
	return fmt.Sprintf("%s", req), nil
}

// GetRequiredData returns error if doesn't have json format
func GetRequiredData(data string) (string, error) {
	jsonBeginIndex := strings.Index(data, "{")
	jsonEndIndex := strings.LastIndex(data, "}")
	if jsonBeginIndex >= jsonEndIndex {
		return "", errors.Errorf("failed to parse the data got: %s", data)
	}
	return data[jsonBeginIndex : jsonEndIndex+1], nil
}

//ServeRequest process the request from the client
func ServeRequest(conn net.Conn, kubeClient *cstorv1alpha1.Kubeclient) {
	var err error
	var readData, filteredData string
	defer func(err error) {
		if err != nil {
			_, er := conn.Write([]byte(respErr + endOfLine))
			if er != nil {
				klog.Errorf("failed to inform to client")
			}
		} else {
			_, er := conn.Write([]byte(respOk + endOfLine))
			if er != nil {
				klog.Errorf("failed to inform to client")
			}
		}
	}(err)
	readData, err = Reader(conn)
	if err != nil {
		klog.Errorf("failed to read data: {%v}", err)
		return
	}
	filteredData, err = GetRequiredData(readData)
	if err != nil {
		klog.Errorf("failed to get required information: {%v}", err)
		return
	}
	replicationData := &cstorv1alpha1.CStorVolumeReplication{}
	err = json.Unmarshal([]byte(filteredData), replicationData)
	if err != nil {
		klog.Errorf("failed to unmarshal replication data {%v}", err)
		return
	}
	csc := &cstorv1alpha1.CStorVolumeConfig{
		CStorVolumeReplication: replicationData,
		Kubeclient:             kubeClient,
	}
	err = csc.UpdateCVWithReplicationDetails()
	if err != nil {
		klog.Errorf("failed to update cstorvolume {%s} with details {%v}"+
			" error: {%v}", csc.VolumeName,
			replicationData, err)
		return
	}
}

// StartTargetServer starts the UnixDomainServer
func StartTargetServer(kubeConfig string) {

	klog.Info("Starting unix domain server")
	if err := os.RemoveAll(string(volumeMgmtUnixSock)); err != nil {
		klog.Fatalf("failed to clear path: {%v}", err)
	}

	listen, err := net.Listen("unix", volumeMgmtUnixSock)
	if err != nil {
		klog.Fatalf("listen error: {%v}", err)
	}

	//TODO: Remove hard coded ENV
	namespace := env.Get("CSTOR_TARGET_NAMESPACE")
	if namespace == "" {
		klog.Fatalf("failed to get volume namespace empty value for env %s",
			"CStorVolumeReplication",
		)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		klog.Fatalf("Caught signal %s: shutting down", sig)
		err := ln.Close()
		if err != nil {
			klog.Errorf("failed to close the connection error: %v", err)
		}
	}(listen, sigc)

	// Since we are reading kubeClient there is no need to taking lock
	kubeClient := cstorv1alpha1.NewKubeclient(
		cstorv1alpha1.WithKubeConfigPath(kubeConfig)).
		WithNamespace(namespace)

	for {
		sockFd, err := listen.Accept()
		if err != nil {
			klog.Fatalf("failed to accept error: {%v}", err)
		}
		go ServeRequest(sockFd, kubeClient)
	}
}
