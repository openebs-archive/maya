# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"
Vagrant.require_version ">= 1.6.0"

# Maya Nodes
M_NODES = ENV['M_NODES'] || 1

# Maya Memory & CPUs
M_MEM = ENV['M_MEM'] || 2048
M_CPUS = ENV['M_CPUS'] || 1

# Generic installer script common for server(s) & client(s)
# This expects arguments that provide runtime values
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
# A bit of conditional logic s.t. we do not repeat CD-ing
grep "cd /opt/gopath/src/github.com/openebs/maya" /home/vagrant/.profile || \
  echo "cd /opt/gopath/src/github.com/openebs/maya" >> /home/vagrant/.profile

echo "In-order to compile maya, look at various options provided in GNUmakefile"
echo -e "\n\tTIP: Start with command:- make"
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
  
  return vmCfg
end

# Entry point of this Vagrantfile
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  # I do not want this
  #config.vbguest.auto_update = false
  
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
