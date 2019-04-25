// Copyright © 2017-2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

// #############################################################################

// Note: The following code is custom settings particular to the new CRD in the
// cluster.

// #############################################################################

// CRDDomain : CRD Domain Name
const CRDDomain string = "openebs.io"

// CRDVersionV1 : CRD Version
const CRDVersionV1 string = "v1"

// CRDSBAName : CRD Storage Backend Adaptor Name
const CRDSBAName string = "storagebackendadaptor"

// CRDSBAResourceName : CRD Storage Backend Adaptor Resource Name
const CRDSBAResourceName string = "storagebackendadaptor"

// CRDSBAResourceNamePlural : CRD Storage Backend Adaptor Resource Name in plural
const CRDSBAResourceNamePlural string = "storagebackendadaptors"
