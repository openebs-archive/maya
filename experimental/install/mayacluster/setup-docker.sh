#!/bin/bash

echo "Provisioning docker ..."

echo "Step: Updating package database"
sudo apt-get update

echo "Step: Add the GPG key for the official Docker repository to the system"
sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 \
  --recv-keys 58118E89F3A912897C070ADBF76221572C52609D

echo "Step: Add the Docker repository to APT sources"
echo "deb https://apt.dockerproject.org/repo ubuntu-xenial main" | \
  sudo tee /etc/apt/sources.list.d/docker.list

echo "Step: Update package database with the Docker packages from the newly added repo"
sudo apt-get update

echo "Step: This makes sure to install from the Docker repo than the default Ubuntu 16.04 repo"
apt-cache policy docker-engine

echo "Step: Install docker"
sudo apt-get install -y docker-engine

echo "Step: Verify if docker is running"
sudo systemctl status docker

echo "Step: Check system-wide info about docker"
docker info

echo "Congratulations docker was provisioned successfully ..."
