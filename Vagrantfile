# -*- mode: ruby -*-
# vi: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"
DEFAULT_CPU_COUNT = 2
DEFAULT_MEMORY = "2048"

$script = <<SCRIPT

# Update apt and get dependencies
#sudo apt-get update
#sudo apt-get install -y zip unzip curl

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

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|

  # I do not want this
  config.vbguest.auto_update = false
  
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
