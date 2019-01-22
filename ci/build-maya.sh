#!/usr/bin/env bash

echo "*****************************Retagging images and setting up env***************************"
#Images from this repo are always tagged as ci
#The downloaded operator file will may contain a non-ci tag name
# depending on when and from where it is being downloaded. For ex:
# - during the release time, the image tags can be versioned like 0.7.0-RC..
# - from a branch, the image tags can be the branch names like v0.7.x-ci
if [ ${CI_TAG} != "ci" ]; then
  sudo docker tag openebs/m-apiserver:ci openebs/m-apiserver:${CI_TAG}
  sudo docker tag openebs/m-exporter:ci openebs/m-exporter:${CI_TAG}
  sudo docker tag openebs/cstor-pool-mgmt:ci openebs/cstor-pool-mgmt:${CI_TAG}
  sudo docker tag openebs/cstor-volume-mgmt:ci openebs/cstor-volume-mgmt:${CI_TAG}
fi

#Tag the images with quay.io, since the operator can either have quay or docker images
sudo docker tag openebs/m-apiserver:ci quay.io/openebs/m-apiserver:${CI_TAG}
sudo docker tag openebs/m-exporter:ci quay.io/openebs/m-exporter:${CI_TAG}
sudo docker tag openebs/cstor-pool-mgmt:ci quay.io/openebs/cstor-pool-mgmt:${CI_TAG}
sudo docker tag openebs/cstor-volume-mgmt:ci quay.io/openebs/cstor-volume-mgmt:${CI_TAG}

## install iscsi pkg
echo "Installing iscsi packages"
sudo apt-get install open-iscsi
sudo service iscsid start
sudo service iscsid status
echo "Installation complete"