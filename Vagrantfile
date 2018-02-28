# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"
Vagrant.require_version ">= 1.6.0"

# Maya Nodes
M_NODES = ENV['M_NODES'] || 1

# Maya Memory & CPUs
M_MEM = ENV['M_MEM'] || 4096
M_CPUS = ENV['M_CPUS'] || 2

# Common installation script
$installer = <<SCRIPT
#!/bin/bash

echo Will run the common installer script ...

# Update apt and get dependencies
sudo apt-get update
sudo apt-get install -y zip unzip curl wget

SCRIPT

$mayascript = <<SCRIPT
#!/bin/bash

cd /opt/gopath/src/github.com/openebs/maya

# Install dependencies required for Development
bash buildscripts/install_go.sh

# CD into the maya working directory when we login to the VM
grep "cd /opt/gopath/src/github.com/openebs/maya" /home/vagrant/.profile || \
  echo "cd /opt/gopath/src/github.com/openebs/maya" >> /home/vagrant/.profile

echo ""
echo "================================================"
echo "Congrats!! Maya has been setup for development"
echo "================================================"
echo ""

SCRIPT

$minikubescript = <<SCRIPT
#!/bin/bash

#Install latest minikube
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
&& chmod +x minikube \
&& sudo mv minikube /usr/local/bin/

#Install latest kubectl
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
&& chmod +x kubectl \
&& sudo mv kubectl /usr/local/bin/

#Setup minikube
mkdir -p $HOME/.minikube
mkdir -p $HOME/.kube
touch $HOME/.kube/config

# Push these ENV k:v to /home/vagrant/.profile to 
# start/restart minikube. This vagrant VM might have been created
# earlier with minikube stopped for some reason.
grep "MINIKUBE_WANTUPDATENOTIFICATION=false" /home/vagrant/.profile || \
  echo "MINIKUBE_WANTUPDATENOTIFICATION=false" >> /home/vagrant/.profile

grep "MINIKUBE_WANTREPORTERRORPROMPT=false" /home/vagrant/.profile || \
  echo "MINIKUBE_WANTREPORTERRORPROMPT=false" >> /home/vagrant/.profile

grep "MINIKUBE_HOME=$HOME" /home/vagrant/.profile || \
  echo "MINIKUBE_HOME=$HOME" >> /home/vagrant/.profile

grep "CHANGE_MINIKUBE_NONE_USER=true" /home/vagrant/.profile || \
  echo "CHANGE_MINIKUBE_NONE_USER=true" >> /home/vagrant/.profile

grep "KUBECONFIG=$HOME/.kube/config" /home/vagrant/.profile || \
  echo "KUBECONFIG=$HOME/.kube/config" >> /home/vagrant/.profile

# Export above as well for `minikube start` to work 
# in the same session of `vagrant up`
export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
export KUBECONFIG=$HOME/.kube/config

# Permissions
sudo chown -R $USER $HOME/.kube
sudo chgrp -R $USER $HOME/.kube

sudo chown -R $USER $HOME/.minikube
sudo chgrp -R $USER $HOME/.minikube

# Start minikube on this host itself
sudo -E minikube start --vm-driver=none --extra-config=apiserver.Authorization.Mode=RBAC

# Wait for Kubernetes to be up and ready.
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until kubectl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done

echo ""
echo "================================================"
echo "Congrats!! minikube apiserver is running"
echo "================================================"
echo ""

# Download and initialize helm.
bash ci/ubuntu-compile-nsenter.sh && sudo cp .tmp/util-linux-2.30.2/nsenter /usr/bin
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get > get_helm.sh
chmod 700 get_helm.sh

bash ./get_helm.sh
helm init

echo ""
echo "================================================"
echo "Congrats!! helm is installed"
echo "================================================"
echo ""

# run the ci
bash ci/travis-ci.sh

SCRIPT

required_plugins = %w(vagrant-cachier)

required_plugins.each do |plugin|
  need_restart = false
  unless Vagrant.has_plugin? plugin
    system "vagrant plugin install #{plugin}"
    need_restart = true
  end
  exec "vagrant #{ARGV.join(' ')}" if need_restart
end


def configureVM(vmCfg, hostname, cpus, mem)

  vmCfg.vm.box = "bento/ubuntu-16.04"
  
  vmCfg.vm.hostname = hostname
  vmCfg.vm.network "private_network", type: "dhcp"

  
  #Adding Vagrant-cachier
  if Vagrant.has_plugin?("vagrant-cachier")
     vmCfg.cache.scope = :machine
     vmCfg.cache.enable :apt
     vmCfg.cache.enable :gem
  end
  
  # Set resources w.r.t Virtualbox provider
  vmCfg.vm.provider "virtualbox" do |vb|
    vb.memory = mem
    vb.cpus = cpus
    vb.customize ["modifyvm", :id, "--cableconnected1", "on"]
  end

  # ensure docker is installed
  vmCfg.vm.provision "docker"

  # sync your laptop's development with this Vagrant VM
  vmCfg.vm.synced_folder '.', '/opt/gopath/src/github.com/openebs/maya'

  # Script to prepare the VM
  vmCfg.vm.provision "shell", inline: $installer, privileged: false 
  vmCfg.vm.provision "shell", inline: $mayascript, privileged: false
  vmCfg.vm.provision "shell", inline: $minikubescript, privileged: false
  
  return vmCfg
end

# Entry point of this Vagrantfile
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  
  # maya master related only !!
  1.upto(M_NODES.to_i) do |i|
    hostname = "maya-%02d" % [i]
    cpus = M_CPUS
    mem = M_MEM
    
    config.vm.define hostname do |vmCfg|
      vmCfg = configureVM(vmCfg, hostname, cpus, mem)
    end     
  end

end
