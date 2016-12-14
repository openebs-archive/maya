#!/bin/bash

set -e

echo "Installing Docker ..."

echo deb https://apt.dockerproject.org/repo ubuntu-`lsb_release -c \
  | awk '{print $2}'` main | sudo tee /etc/apt/sources.list.d/docker.list

# TODO - These needs to be coming from maya cli
sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
sudo apt-get install -y docker-engine

# Restart docker to make sure we get the latest version of the daemon
# if there is an upgrade
sudo service docker restart

# Make sure we can actually use docker as the vagrant user
# TODO - Need to think about user
#sudo usermod -aG docker vagrant
