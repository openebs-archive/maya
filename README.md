[![Build Status](https://app.travis-ci.com/openebs/maya.svg?branch=v2.12.x)](https://app.travis-ci.com/openebs/maya)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/maya)](https://goreportcard.com/report/github.com/openebs/maya)
[![codecov](https://codecov.io/gh/openebs/maya/branch/v2.12.x/graph/badge.svg?token=nDwloue1T5)](https://codecov.io/gh/openebs/maya)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/openebs/maya/blob/HEAD/LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fmaya.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fmaya?ref=badge_shield)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/1753/badge)](https://bestpractices.coreinfrastructure.org/projects/1753)

## Overview

OpenEBS control plane components like provisioners and operators were hosted in this repository. 

As the OpenEBS community started to add new engines, the engine specific control plane components have been moved to their respective repositories.
- cStor operators and CSI driver have moved to [openebs/cstor-operators](https://github.com/openebs/cstor-operators) and [openebs/cstor-csi](https://github.com/openebs/cstor-csi) respectively.
- Jiva operator and CSI driver have moved to [openebs/jiva-operator](https://github.com/openebs/jiva-operator)
- Local PV Hostpath and Device provisioner has been moved to [openebs/dynamic-localpv-provisioner](https://github.com/openebs/dynamic-localpv-provisioner)
- Jiva and cStor prometheus metrics exporter been moved to [openebs/m-exporter](https://github.com/openebs/m-exporter)
- `mayactl` for displaying status of Jiva and cStor volumes is merged into [openebs/openebsctl](https://github.com/openebs/openebsctl)

This repository mainly contains code required for running the legacy cStor and Jiva pools and volumes like: 
- `m-apiserver` - used for provisoining the legacy cStor and Jiva pools and volumes.
- `mayactl` - packaged along with `m-apiserver` for fetching the legacy cStor and Jiva volume status. 
- `admission-server` - used for validating Jiva and cStor pool and volume requests. 
- `m-upgrade` - used for upgrading the legacy Jiva volumes, cStor pools and volumes.
- `cstor-pool-mgmt` and `cstor-volume-mgmt` - used for managing the legacy cStor pool and volumes. 

With OpenEBS 3.0, all of the above legacy components are deprecated and users are requested to migrate towards using:
- CStor CSI Driver
- Jiva CSI Driver

The steps to migrate are provided here: https://github.com/openebs/upgrade.

`v2.12.x` is the last active branch on this repository, that will be used to mainly resolve any security vulnerability or kubernetes compatibility issues found on production setups using the legacy provisioners. New features will be developed only cStor and Jiva CSI drivers.

## Install

Please refer to our documentation at [OpenEBS Documentation](http://openebs.io/).

## Release

Prior to creating a release tag on this repository on `v2.12.x` branch with the required fixes, ensure that the dependent data engine repositories and provisioner are tagged. Once the code is merged, use the following sequence to release a new version for the legacy components:
- (Optional) New release tag on v2.12.x branch of [openebs/linux-utils](https://github.com/openebs/linux-utils)
- (Optional) New release tag on v0.6.x branch of [openebs/ndm](https://github.com/openebs/node-disk-manager)
- New release tag on v2.12.x branch of [openebs/cstor](https://github.com/openebs/cstor) and [openebs/libcstor](https://github.com/openebs/libcstor)
- New release tag on v2.12.x branch of [openebs/jiva](https://github.com/openebs/jiva)
- New release tag on v2.12.x branch of [openebs/openebs-k8s-provisioner](https://github.com/openebs/openebs-k8s-provisioner) 
- New release tag on v2.12.x branch of [openebs/m-exporter](https://github.com/openebs/m-exporter)
- New release tag on v2.12.x branch of [openebs/maya](https://github.com/openebs/maya)
- New release tag on v2.12.x branch of [openebs/velero-plugin](https://github.com/openebs/velero-plugin)

## Contributing

We are looking at further refactoring this repository by moving the common packages from this repository into a new common repository. If you are interested in helping with the refactoring efforts, please reach out to the OpenEBS Community. 

For details on setting up the development environment and fixing the code, head over to the [CONTRIBUTING.md](./CONTRIBUTING.md).

## Community

OpenEBS welcomes your feedback and contributions in any form possible.

- [Join OpenEBS community on Kubernetes Slack](https://kubernetes.slack.com)
  - Already signed up? Head to our discussions at:
    -  [#openebs](https://kubernetes.slack.com/messages/openebs/)
    -  [#openebs-dev](https://kubernetes.slack.com/messages/openebs-dev/)

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Fmaya.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Fmaya?ref=badge_large)
