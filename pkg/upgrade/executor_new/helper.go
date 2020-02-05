package executor

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
