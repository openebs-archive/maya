# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"
Vagrant.require_version ">= 1.6.0"

# Server Nodes
S_NODES = ENV['S_NODES'] || 1

# Client Nodes
C_NODES = ENV['C_NODES'] || 2

# Server Memory & CPUs
S_MEM = ENV['S_MEM'] || 512
S_CPUS = ENV['S_CPUS'] || 1

# Client Memory & CPUs
C_MEM = ENV['C_MEM'] || 1024
C_CPUS = ENV['C_CPUS'] || 1

# Private IP address prefix of server(s) & client(s)
BASE_SIP_ADDR = ENV['BASE_SIP_ADDR'] || "10.21.0"
BASE_CIP_ADDR = ENV['BASE_CIP_ADDR'] || "10.31.0"

$maya = <<SCRIPT

# Update apt and get dependencies
sudo apt-get install -y zip

cd /opt/gopath/src/github.com/openebs/maya

# Install dependencies required for Development
bash scripts/install_go.sh

# CD into the maya working directory when we login to the VM
# A bit of conditional logic s.t. we do not repeat CD-ing
grep "cd /opt/gopath/src/github.com/openebs/maya" ~/.profile || \
  echo "cd /opt/gopath/src/github.com/openebs/maya" >> ~/.profile

echo "In-order to compile maya, look at various options provided in GNUmakefile"
echo -e "\n\tTIP: Start with command:- make bootstrap"
SCRIPT

$serverpeers = <<SCRIPT
#!/bin/bash

if [ $# -ne 1 ]; then
    echo "usage: $0 ALL_SERVER_IPS"
    exit 1
fi

ALL_SERVER_IPS=$1

cd /opt/gopath/src/github.com/openebs/maya

echo "${ALL_SERVER_IPS}" > ./serverpeers
SCRIPT


$clientpeers = <<SCRIPT
#!/bin/bash

if [ $# -ne 1 ]; then
    echo "usage: $0 ALL_CLIENT_IPS"
    exit 1
fi

ALL_CLIENT_IPS=$1

cd /opt/gopath/src/github.com/openebs/maya

echo "${ALL_CLIENT_IPS}" > ./clientpeers
SCRIPT

def configureVM(vmCfg, hostname, cpus, mem, self_ipv4, all_servers_ipv4, all_clients_ipv4)

  vmCfg.vm.box = "bento/ubuntu-16.04"
  
  vmCfg.vm.hostname = hostname
  
  # Set resources w.r.t Virtualbox provider
  vmCfg.vm.provider "virtualbox" do |vb|
    vb.memory = mem
    vb.cpus = cpus
    vb.customize ["modifyvm", :id, "--cableconnected1", "on"]
  end
  
  vmCfg.vm.network "private_network", ip: self_ipv4 
  
  # sync your laptop's development with this Vagrant VM
  vmCfg.vm.synced_folder '.', '/opt/gopath/src/github.com/openebs/maya'
  
  vmCfg.vm.provision "shell", inline: $maya, privileged: false    
        
  vmCfg.vm.provision "shell" do |s|
    s.inline = $serverpeers
    s.privileged = false
    s.args = [all_servers_ipv4]
  end
  
  
  vmCfg.vm.provision "shell" do |s|
    s.inline = $clientpeers
    s.privileged = false
    s.args = [all_clients_ipv4]
  end
  
  return vmCfg
end

def updateServerIPs(all_servers_ipv4, self_ipv4)

    if all_servers_ipv4.to_s != ''
      all_servers_ipv4 = all_servers_ipv4 + ", "    
    end
    
    all_servers_ipv4 = all_servers_ipv4 + '"' + self_ipv4 + '"'
    
    return all_servers_ipv4
end

def updateClientIPs(all_clients_ipv4, self_ipv4)

    if all_clients_ipv4.to_s != ''
      all_clients_ipv4 = all_clients_ipv4 + ", "    
    end
    
    all_clients_ipv4 = all_clients_ipv4 + '"' + self_ipv4 + '"'
    
    return all_clients_ipv4
end

# Entry point of this Vagrantfile
Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  # I do not want this
  config.vbguest.auto_update = false
  
  # Placeholder to store comma separated ip addresses
  all_servers_ipv4 = ""
  all_clients_ipv4 = ""

  1.upto(S_NODES.to_i) do |i|
    self_ipv4 = "#{BASE_SIP_ADDR}.#{i+99}"
    all_servers_ipv4 = updateServerIPs(all_servers_ipv4, self_ipv4)
  end
  
  1.upto(C_NODES.to_i) do |i|
    self_ipv4 = "#{BASE_CIP_ADDR}.#{i+99}"
    all_clients_ipv4 = updateClientIPs(all_clients_ipv4, self_ipv4)
  end
  
  # Server related only !!
  1.upto(S_NODES.to_i) do |i|
    hostname = "server-%02d" % [i]
    cpus = S_CPUS
    mem = S_MEM
    
    self_ipv4 = "#{BASE_SIP_ADDR}.#{i+99}"
    
    config.vm.define hostname do |vmCfg|
      vmCfg = configureVM(vmCfg, hostname, cpus, mem, self_ipv4, all_servers_ipv4, all_clients_ipv4)
    end    
  end
  
  # Client related only !!
  1.upto(C_NODES.to_i) do |i|
    hostname = "client-%02d" % [i]
    cpus = C_CPUS
    mem = C_MEM
    
    self_ipv4 = "#{BASE_CIP_ADDR}.#{i+99}"
    
    config.vm.define hostname do |vmCfg|
      vmCfg = configureVM(vmCfg, hostname, cpus, mem, self_ipv4, all_servers_ipv4, all_clients_ipv4)
    end
  end

end
