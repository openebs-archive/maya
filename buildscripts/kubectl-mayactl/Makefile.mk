# Copyright © 2017 The OpenEBS Authors
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

# Specify the name for the binaries
MAYACTL=kubectl-mayactl

.PHONY: mayactl
mayactl:
	@echo "----------------------------"
	@echo "--> mayactl                    "
	@echo "----------------------------"
	@PNAME="kubectl-mayactl" CTLNAME=${MAYACTL} sh -c "'$(PWD)/buildscripts/build.sh'"
	@echo "--> Removing old directory..."
	@sudo rm -rf /usr/local/bin/${MAYACTL}
	@echo "----------------------------"
	@echo "copying new mayactl"
	@echo "----------------------------"
	@sudo mkdir -p  /usr/local/bin/
	@sudo cp -a "$(GOPATH)/bin/${MAYACTL}"  /usr/local/bin/${MAYACTL}
	@sudo rm "$(GOPATH)/bin/${MAYACTL}"
	@echo "=> copied to /usr/local/bin"