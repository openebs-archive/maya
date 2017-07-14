# Maya API Server

[![GoDoc](https://godoc.org/github.com/openebs/mayaserver?status.png)](https://godoc.org/github.com/openebs/mayaserver) [![Build Status](https://travis-ci.org/openebs/mayaserver.svg?branch=master)](https://travis-ci.org/openebs/mayaserver) [![Go Report Card](https://goreportcard.com/badge/github.com/openebs/mayaserver)](https://goreportcard.com/report/github.com/openebs/mayaserver) [![codecov](https://codecov.io/gh/openebs/mayaserver/branch/master/graph/badge.svg)](https://codecov.io/gh/openebs/mayaserver)

> OpenEBS exposes its APIs here

A service exposing `Kubernetes` like volume APIs.

## Setting up maya api server development & run environment

> These are some of the steps to start off with development & running of maya api 
server in one's laptop. It assumes use of Linux as laptop's OS. In addition, the 
laptop should have Virtual Box & Vagrant installed.

```bash
- git clone https://github.com/openebs/mayaserver.git
- cd to above cloned folder i.e mayaserver
  - vagrant up
  - vagrant ssh
- Inside the vagrant VM run below steps:
  - make init
  - make
  - make bin
  - nohup m-apiserver up -bind=172.28.128.4 &>mapiserver.log &
```

### Troubleshooting during local setup

```bash
- `make init` is a time taking operation
  - This downloads all the vendoring libraries
  - Typically required for the very first attempt only
  - In case of add/update of new/existing vendoring libraries:
    - use `make sync` than `make init`
```

## Installing maya api server

> Steps to install maya api server's released binary

```bash
- Navigate to https://github.com/openebs/mayaserver/releases
- Download the binary depending on the required release & host OS architecture
  - e.g. Below link points to release `0.0.6` & `linux_386` architecture
  - https://github.com/openebs/mayaserver/releases/download/0.0.6/mayaserver-linux_386.zip
- Extract the m-apiserver binary from above .zip file & put it at /usr/local/bin/
- Set appropriate user groups & permissions
- Start maya api server as a long running daemon
  - sudo nohup m-apiserver up -bind=<IP_Addr> &>mapiserver.log &
```

## Troubleshooting

- Verify the presence of maya api server binary
  - which m-apiserver
  - m-apiserver version

- Verify if maya api server is running as a process
  - Watch out for the process with 5656 as the port
  - `5656` is the default tcp port on which maya api server's services are exposed

  ```bash
  # Use netstat command
  $ netstat -tnlp

  (Not all processes could be identified, non-owned process info
   will not be shown, you would have to be root to see it all.)
  Active Internet connections (only servers)
  Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
  tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      -
  tcp        0      0 127.0.0.1:5656          0.0.0.0:*               LISTEN      -
  tcp6       0      0 :::22                   :::*                    LISTEN      -

  # Using sudo will display the PID details
  $ sudo netstat -tnlp

  Active Internet connections (only servers)
  Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
  tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      1258/sshd
  tcp        0      0 127.0.0.1:5656          0.0.0.0:*               LISTEN      3078/m-apiserver 
  tcp6       0      0 :::22                   :::*                    LISTEN      1258/sshd
  ```

## Licensing

Maya api server is completely open source and bears an Apache license. Maya api 
server's core components and designs are a derivative of other open sourced libraries 
like Nomad and Kubernetes.
