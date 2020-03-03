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

package upgrader

import (
	"strings"

	"github.com/pkg/errors"
)

func getImageURL(url, prefix string) (string, error) {
	lastIndex := strings.LastIndex(url, ":")
	if lastIndex == -1 {
		return "", errors.Errorf("no version tag found on image %s", url)
	}
	baseImage := url[:lastIndex]
	if prefix != "" {
		// urlPrefix is the url to the directory where the images are present
		// the below logic takes the image name from current baseImage and
		// appends it to the given urlPrefix
		// For example baseImage is abc/quay.io/openebs/jiva
		// and urlPrefix is xyz/aws-56546546/openebsdirectory/
		// it will take jiva from current url and append it to urlPrefix
		// and return xyz/aws-56546546/openebsdirectory/jiva
		urlSubstr := strings.Split(baseImage, "/")
		baseImage = prefix + urlSubstr[len(urlSubstr)-1]
	}
	return baseImage, nil
}
