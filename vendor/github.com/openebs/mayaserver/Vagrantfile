# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"
Vagrant.require_version ">= 1.6.0"

# Mayaserver Nodes
NODES = ENV['NODES'] || 1

# Memory & CPUs
MEM = ENV['MEM'] || 2048
CPUS = ENV['CPUS'] || 2

# Common installation stuff
$installer = <<SCRIPT
#!/bin/bash

echo Will run the common installer script ...

# Update apt and get dependencies
sudo apt-get update
sudo apt-get install -y zip unzip curl wget

SCRIPT

$mayaserverdev = <<SCRIPT
#!/bin/sh

cd /opt/gopath/src/github.com/openebs/mayaserver

# Install dependencies required for Development
bash buildscripts/install_go.sh

# CD into the mayaserver working directory when we login to the VM
# A bit of conditional logic s.t. we do not repeat CD-ing
grep "cd /opt/gopath/src/github.com/openebs/mayaserver" /home/vagrant/.profile || \
  echo "cd /opt/gopath/src/github.com/openebs/mayaserver" >> /home/vagrant/.profile

echo "Look into the GNUmakefile & invoke init to get started with development"
SCRIPT

required_plugins = %w(vagrant-cachier vagrant-vbguest)

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
  
  # sync your laptop's development with this Vagrant VM
  vmCfg.vm.synced_folder '.', '/opt/gopath/src/github.com/openebs/mayaserver'

  # Script will make some directories before installation procedure
  vmCfg.vm.provision "shell", inline: $installer, privileged: false 
  vmCfg.vm.provision "shell", inline: $mayaserverdev, privileged: false
  vmCfg.vm.provision "docker" # Just install it
  
  return vmCfg
end

# Entry point of this Vagrantfile
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  # I do not want this
  config.vbguest.auto_update = false
  
  1.upto(NODES.to_i) do |i|
    hostname = "maya-dev-node-%02d" % [i]
    cpus = CPUS
    mem = MEM
    
    config.vm.define hostname do |vmCfg|
      vmCfg = configureVM(vmCfg, hostname, cpus, mem)  
    end
  end

end
