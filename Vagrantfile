# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"
DEFAULT_CPU_COUNT = 2
DEFAULT_MEMORY = "2048"

$script = <<SCRIPT

# Update apt and get dependencies
sudo apt-get update
sudo apt-get install -y zip unzip curl

cd /opt/gopath/src/github.com/openebs/maya

# Install dependencies required for 
# 1/ Running functional tests
# 2/ Development (linting, coverage analysis, etc.)
bash scripts/install_go.sh
bash scripts/install_nomad.sh
bash scripts/install_docker.sh

# Start Nomad in dev mode
nohup nomad agent -dev -config=scripts/config &>nomad.log  &

# CD into the maya working directory when we login to the VM
# A bit of conditional logic s.t. we do not repeat CD-ing
grep "cd /opt/gopath/src/github.com/openebs/maya" ~/.profile || \
  echo "cd /opt/gopath/src/github.com/openebs/maya" >> ~/.profile
  
echo "In-order to run maya, look at various options provided in GNUmakefile"

SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  vmName = "maya-dev"
  config.vm.box = "bento/ubuntu-16.04"

  config.vm.define vmName do |vmCfg|
      vmCfg.vm.hostname = vmName

      # sync your laptop's development with this Vagrant VM
      vmCfg.vm.synced_folder '.', '/opt/gopath/src/github.com/openebs/maya'
      
      vmCfg.vm.provision "shell", inline: $script, privileged: false
      
      vmCfg.vm.provider "virtualbox" do |vb|
        vb.memory = DEFAULT_MEMORY
  	    vb.cpus = DEFAULT_CPU_COUNT
  	    vb.customize ["modifyvm", :id, "--cableconnected1", "on"]
      end
  end

end
