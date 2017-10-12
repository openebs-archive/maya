## Overview

[![Build Status](https://travis-ci.org/openebs/maya.svg?branch=master)](https://travis-ci.org/openebs/maya) 
[![Go Report](https://goreportcard.com/badge/github.com/openebs/maya)](https://goreportcard.com/report/github.com/openebs/maya) [![codecov](https://codecov.io/gh/openebs/maya/branch/master/graph/badge.svg)](https://codecov.io/gh/openebs/maya) [![GoDoc](https://godoc.org/github.com/openebs/maya?status.svg)](https://godoc.org/github.com/openebs/maya)

*This repository hosts the source code for OpenEBS Storage Orchestration Engine.* 

OpenEBS Storage Orchestration (aka *Maya*, meaning *Magic*) aims at managing storage for millions of containers with deceptive simplicity. Maya is the control plane for OpenEBS Storage. Maya is developed with a focused *goal* of easing(and at times completely eliminating) the Storage Ops for containers like 
- managing Storage Volumes (Persistent Volumes) and 
- managing Storage Infrastructure (Physical Storage - Local/Cloud Disks/SSDs/Cache, Controllers, Networks). 

OpenEBS Maya allows you to manage storage across multiple zones (aka clusters/environments), that are co-located or geographically separated and can also run from within a single host. Maya can move the storage across different tiers based on the application needs (volume migration). Maya can learn and adapt to the changing environment through machine learning and data analytics. 

## Design

Maya - *Storage Orchestration* functionality is delivered through a set of services and tools that seamlessly integrate *Maya* into your Container Orchestrators like Kubernetes, Docker Swarm, etc.,  Maya comprises of several components, which are themself delivered as container images. 

**Maya** components can be broadly classified based on their deployment type as follows:

- **Cluster Components** - like API Server (maya-apiserver) that helps in processing the requests for creating new OpenEBS Volumes. The API server, will use the container orchestrators to provision the OpenEBS Volume Containers. Another example is the Smart Analytics (maya-mulebot) engine that gathers the data via the machine learning probes and runs heuristics analysis to optimize storage deployment. 

- **Node/Host Components** - are the services and tools that run on each of the nodes or docker hosts. In case of Kubernetes, these components are deployed as DaemonSets. The functionality of these components is local to the node like, Storage Agents (maya-agent) managing the disks attached to the hosts and helping in carving out the required disks to different OpenEBS Volumes. The agent can interact with Cluster components and vice-versa. 

Maya makes the storage infrastructure programmable via the yaml files. 

## To start using Maya

Please refer to our documentation at [OpenEBS Documentation](http://openebs.readthedocs.io/en/latest/)

## Start developing Maya

Head over to the [developer's documentation](https://github.com/openebs/maya/blob/master/docs/developer.md) for more details.

### Dependency management
Maya uses [golang dep] to manage dependencies. Usage can be found on [dep README].

[Go environment]: https://golang.org/doc/install
[developer's documentation]: https://github.com/openebs/maya/blob/master/docs/developer.md
[golang dep]: https://github.com/golang/dep
[dep README]: https://github.com/golang/dep#usage
