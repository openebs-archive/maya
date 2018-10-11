# Copyright 2018 The OpenEBS Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# openebs rhel base image used for image security scan
FROM openebs/rhel7

ENV MAYA_API_SERVER_NETWORK="eth0"

COPY maya-apiserver /usr/local/bin/
COPY mayactl /usr/local/bin/
COPY licenses /licenses
COPY entrypoint.sh /usr/local/bin/

ARG VERSION
LABEL version=$VERSION
LABEL com.redhat.component="storage"
LABEL name="openebs/m-apiserver"
LABEL description="OpenEBS storage API server"
LABEL summary=" OpenEBS API Server with RHEL, used for managing OpenEBS Volumes"
LABEL io.k8s.display-name="OpenEBS APISERVER"
LABEL io.openshift.tags="openebs apiserver storage"
LABEL License="Apache-2.0"
LABEL release=$VERSION
LABEL vendor="OpenEBS"

RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT entrypoint.sh "${MAYA_API_SERVER_NETWORK}"
