# -*- mode: ruby -*-
# vi: set ft=ruby :

require 'fileutils'
require 'net/http'
require 'open-uri'
require 'json'

class Module
  def redefine_const(name, value)
    __send__(:remove_const, name) if const_defined?(name)
    const_set(name, value)
  end
end

module OS
  def OS.windows?
    (/cygwin|mswin|mingw|bccwin|wince|emx/ =~ RUBY_PLATFORM) != nil
  end

  def OS.mac?
   (/darwin/ =~ RUBY_PLATFORM) != nil
  end

  def OS.unix?
    !OS.windows?
  end

  def OS.linux?
    OS.unix? and not OS.mac?
  end
end

required_plugins = %w(vagrant-triggers)

# check either 'http_proxy' or 'HTTP_PROXY' environment variable
enable_proxy = !(ENV['HTTP_PROXY'] || ENV['http_proxy'] || '').empty?
if enable_proxy
  required_plugins.push('vagrant-proxyconf')
end

if OS.windows?
  puts "You're running an unsupported platform. Exiting.."
  exit
end

required_plugins.push('vagrant-timezone')

required_plugins.each do |plugin|
  need_restart = false
  unless Vagrant.has_plugin? plugin
    system "vagrant plugin install #{plugin}"
    need_restart = true
  end
  exec "vagrant #{ARGV.join(' ')}" if need_restart
end

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"
Vagrant.require_version ">= 1.6.0"

# cloud-configs
SERVER_YAML = File.join(File.dirname(__FILE__), "server.yaml")
CLIENT_YAML = File.join(File.dirname(__FILE__), "client.yaml")

USE_DOCKERCFG = ENV['USE_DOCKERCFG'] || false
DOCKERCFG = File.expand_path(ENV['DOCKERCFG'] || "~/.dockercfg")

DOCKER_OPTIONS = ENV['DOCKER_OPTIONS'] || ''

NOMAD_VERSION = ENV['NOMAD_VERSION'] || '0.4.1'
CONSUL_VERSION = ENV['CONSUL_VERSION'] || '0.6.4'

CHANNEL = ENV['CHANNEL'] || 'alpha'
#if CHANNEL != 'alpha'
#  puts "============================================================================="
#  puts "As this is a fastly evolving technology CoreOS' alpha channel is the only one"
#  puts "expected to behave reliably. While one can invoke the beta or stable channels"
#  puts "please be aware that your mileage may vary a whole lot."
#  puts "So, before submitting a bug, in this project, or upstreams (either Nomad"
#  puts "or CoreOS) please make sure it (also) happens in the (default) alpha channel."
#  puts "============================================================================="
#end

COREOS_VERSION = ENV['COREOS_VERSION'] || 'latest'
upstream = "http://#{CHANNEL}.release.core-os.net/amd64-usr/#{COREOS_VERSION}"
if COREOS_VERSION == "latest"
  upstream = "http://#{CHANNEL}.release.core-os.net/amd64-usr/current"
  url = "#{upstream}/version.txt"
  Object.redefine_const(:COREOS_VERSION,
    open(url).read().scan(/COREOS_VERSION=.*/)[0].gsub('COREOS_VERSION=', ''))
end

NODES = ENV['NODES'] || 2

SERVER_MEM = ENV['SERVER_MEM'] || 512
SERVER_CPU = ENV['SERVER_CPU'] || 1

CLIENT_MEM= ENV['CLIENT_MEM'] || 2048
CLIENT_CPUS = ENV['CLIENT_CPUS'] || 1

BASE_IP_ADDR = ENV['BASE_IP_ADDR'] || "172.17.9"

DNS_DOMAIN = ENV['DNS_DOMAIN'] || "cluster.local"
DNS_UPSTREAM_SERVERS = ENV['DNS_UPSTREAM_SERVERS'] || "8.8.8.8:53,8.8.4.4:53"

SERIAL_LOGGING = (ENV['SERIAL_LOGGING'].to_s.downcase == 'true')
GUI = (ENV['GUI'].to_s.downcase == 'true')
BOX_TIMEOUT_COUNT = ENV['BOX_TIMEOUT_COUNT'] || 50

if enable_proxy
  HTTP_PROXY = ENV['HTTP_PROXY'] || ENV['http_proxy']
  HTTPS_PROXY = ENV['HTTPS_PROXY'] || ENV['https_proxy']
  NO_PROXY = ENV['NO_PROXY'] || ENV['no_proxy'] || "localhost"
end

REMOVE_VAGRANTFILE_USER_DATA_BEFORE_HALT = (ENV['REMOVE_VAGRANTFILE_USER_DATA_BEFORE_HALT'].to_s.downcase == 'true')
# if this is set true, remember to use --provision when executing vagrant up / reload

# Read YAML file with mountpoint details
MOUNT_POINTS = YAML::load_file('synced_folders.yaml')

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  # always use host timezone in VMs
  config.timezone.value = :host
 
  # always use Vagrants' insecure key
  config.ssh.insert_key = false
  config.ssh.forward_agent = true

  config.vm.box = "coreos-#{CHANNEL}"
  config.vm.box_version = ">= #{COREOS_VERSION}"
  config.vm.box_url = "#{upstream}/coreos_production_vagrant.json"

  ["vmware_fusion", "vmware_workstation"].each do |vmware|
    config.vm.provider vmware do |v, override|
      override.vm.box_url = "#{upstream}/coreos_production_vagrant_vmware_fusion.json"
    end
  end

  config.vm.provider :parallels do |vb, override|
    override.vm.box = "AntonioMeireles/coreos-#{CHANNEL}"
    override.vm.box_url = "https://vagrantcloud.com/AntonioMeireles/coreos-#{CHANNEL}"
  end

  config.vm.provider :virtualbox do |v|
    # On VirtualBox, we don't have guest additions or a functional vboxsf
    # in CoreOS, so tell Vagrant that so it can be smarter.
    v.check_guest_additions = false
    v.functional_vboxsf     = false
  end
  config.vm.provider :parallels do |p|
    p.update_guest_tools = false
    p.check_guest_tools = false
  end

  # plugin conflict
  if Vagrant.has_plugin?("vagrant-vbguest") then
    config.vbguest.auto_update = false
  end

  # setup VM proxy to system proxy environment
  if Vagrant.has_plugin?("vagrant-proxyconf") && enable_proxy
    config.proxy.http = HTTP_PROXY
    config.proxy.https = HTTPS_PROXY
    # most http tools, like wget and curl do not undestand IP range
    # thus adding each node one by one to no_proxy
    no_proxies = NO_PROXY.split(",")
    (1..(NODES.to_i + 1)).each do |i|
      vm_ip_addr = "#{BASE_IP_ADDR}.#{i+100}"
      Object.redefine_const(:NO_PROXY,
        "#{NO_PROXY},#{vm_ip_addr}") unless no_proxies.include?(vm_ip_addr)
    end
    config.proxy.no_proxy = NO_PROXY
    # proxyconf plugin use wrong approach to set Docker proxy for CoreOS
    # force proxyconf to skip Docker proxy setup
    config.proxy.enabled = { docker: false }
  end

  (1..(NODES.to_i + 1)).each do |i|
    if i == 1
      hostname = "server"
      MASTER_IP="#{BASE_IP_ADDR}.#{i+99}"
      ETCD_SEED_CLUSTER = "#{hostname}=http://#{MASTER_IP}:2380"
      cfg = SERVER_YAML
      memory = SERVER_MEM
      cpus = SERVER_CPU
    else
      hostname = "client-%02d" % (i - 1)
      cfg = CLIENT_YAML
      memory = CLIENT_MEM
      cpus = CLIENT_CPUS
    end

    config.vm.define vmName = hostname do |kHost|
      kHost.vm.hostname = vmName

      # suspend / resume is hard to be properly supported because we have no way
      # to assure the fully deterministic behavior of whatever is inside the VMs
      # when faced with XXL clock gaps... so we just disable this functionality.
      kHost.trigger.reject [:suspend, :resume] do
        info "'vagrant suspend' and 'vagrant resume' are disabled."
        info "- please do use 'vagrant halt' and 'vagrant up' instead."
      end

      config.trigger.instead_of :reload do
        exec "vagrant halt && vagrant up"
        exit
      end

      # vagrant-triggers has no concept of global triggers so to avoid having
      # then to run as many times as the total number of VMs we only call them
      # in the server (re: emyl/vagrant-triggers#13)...
      if vmName == "server"
        kHost.trigger.before [:up, :provision] do
          info "Setting up Nomad #{NOMAD_VERSION}"
          sedInplaceArg = OS.mac? ? " ''" : ""
          system "cp setup.tmpl temp/setup"
          system "sed -e 's|__NOMAD_VERSION__|#{NOMAD_VERSION}|g' -i#{sedInplaceArg} ./temp/setup"
          system "sed -e 's|__MASTER_IP__|#{MASTER_IP}|g' -i#{sedInplaceArg} ./temp/setup"
          if enable_proxy
            system "sed -e 's|__PROXY_LINE__||g' -i#{sedInplaceArg} ./temp/setup"
            system "sed -e 's|__NO_PROXY__|#{NO_PROXY}|g' -i#{sedInplaceArg} ./temp/setup"
          else
            system "sed -e '/__PROXY_LINE__/d' -i#{sedInplaceArg} ./temp/setup"
          end
          system "chmod +x temp/setup"
      end

        kHost.trigger.after [:up, :resume] do
          info "Sanitizing stuff..."
          system "ssh-add ~/.vagrant.d/insecure_private_key"
          system "rm -rf ~/.fleetctl/known_hosts"
        end

        kHost.trigger.after [:up] do
          info "Waiting for Nomad server to become ready..."
          j, uri, res = 0, URI("http://#{MASTER_IP}:4646"), nil
          loop do
            j += 1
            begin
              res = Net::HTTP.get_response(uri)
            rescue
              sleep 10
            end
            break if res.is_a? Net::HTTPSuccess or j >= BOX_TIMEOUT_COUNT
          end

          info "Installing nomad CLI for the Nomad version we just bootstrapped..."
          system "./temp/setup install"
        end
      end

      if vmName == "client-%02d" % (i - 1)
        kHost.trigger.after [:up] do
          info "Waiting for Nomad client [client-%02d" % (i - 1) + "] to become ready..."
          j, uri, hasResponse = 0, URI("http://#{BASE_IP_ADDR}.#{i+99}:4646"), false
          loop do
            j += 1
            begin
              res = Net::HTTP.get_response(uri)
              hasResponse = true
            rescue Net::HTTPBadResponse
              hasResponse = true
            rescue
              sleep 10
            end
            break if hasResponse or j >= BOX_TIMEOUT_COUNT
          end
        end
      end

      kHost.trigger.before [:halt, :reload] do
        if REMOVE_VAGRANTFILE_USER_DATA_BEFORE_HALT
          run_remote "sudo rm -f /var/lib/coreos-vagrant/vagrantfile-user-data"
        end
      end

      kHost.trigger.before [:destroy] do
        system <<-EOT.prepend("\n\n") + "\n"
          rm -f temp/*
        EOT
      end

      if SERIAL_LOGGING
        logdir = File.join(File.dirname(__FILE__), "log")
        FileUtils.mkdir_p(logdir)

        serialFile = File.join(logdir, "#{vmName}-serial.txt")
        FileUtils.touch(serialFile)

        ["vmware_fusion", "vmware_workstation"].each do |vmware|
          kHost.vm.provider vmware do |v, override|
            v.vmx["serial0.present"] = "TRUE"
            v.vmx["serial0.fileType"] = "file"
            v.vmx["serial0.fileName"] = serialFile
            v.vmx["serial0.tryNoRxLoss"] = "FALSE"
          end
        end
        kHost.vm.provider :virtualbox do |vb, override|
          vb.customize ["modifyvm", :id, "--uart1", "0x3F8", "4"]
          vb.customize ["modifyvm", :id, "--uartmode1", serialFile]
        end
        # supported since vagrant-parallels 1.3.7
        # https://github.com/Parallels/vagrant-parallels/issues/164
        kHost.vm.provider :parallels do |v|
          v.customize("post-import",
            ["set", :id, "--device-add", "serial", "--output", serialFile])
          v.customize("pre-boot",
            ["set", :id, "--device-set", "serial0", "--output", serialFile])
        end
      end

      ["vmware_fusion", "vmware_workstation", "virtualbox"].each do |h|
        kHost.vm.provider h do |vb|
          vb.gui = GUI
        end
      end
      ["vmware_fusion", "vmware_workstation"].each do |h|
        kHost.vm.provider h do |v|
          v.vmx["memsize"] = memory
          v.vmx["numvcpus"] = cpus
        end
      end
      ["parallels", "virtualbox"].each do |h|
        kHost.vm.provider h do |n|
          n.memory = memory
          n.cpus = cpus
        end
      end

      kHost.vm.network :private_network, ip: "#{BASE_IP_ADDR}.#{i+99}"
      # you can override this in synced_folders.yaml
      kHost.vm.synced_folder ".", "/vagrant", disabled: true

      begin
        MOUNT_POINTS.each do |mount|
          mount_options = ""
          disabled = false
          nfs =  true
          if mount['mount_options']
            mount_options = mount['mount_options']
          end
          if mount['disabled']
            disabled = mount['disabled']
          end
          if mount['nfs']
            nfs = mount['nfs']
          end
          if File.exist?(File.expand_path("#{mount['source']}"))
            if mount['destination']
              kHost.vm.synced_folder "#{mount['source']}", "#{mount['destination']}",
                id: "#{mount['name']}",
                disabled: disabled,
                mount_options: ["#{mount_options}"],
                nfs: nfs
            end
          end
        end
      rescue
      end

      if USE_DOCKERCFG && File.exist?(DOCKERCFG)
        kHost.vm.provision :file, run: "always",
         :source => "#{DOCKERCFG}", :destination => "/home/core/.dockercfg"

        kHost.vm.provision :shell, run: "always" do |s|
          s.inline = "cp /home/core/.dockercfg /root/.dockercfg"
          s.privileged = true
        end
      end

      if File.exist?(cfg)
        kHost.vm.provision :file, :source => "#{cfg}", :destination => "/tmp/vagrantfile-user-data"
        if enable_proxy
          kHost.vm.provision :shell, :privileged => true,
          inline: <<-EOF
          sed -i"*" "s|__PROXY_LINE__||g" /tmp/vagrantfile-user-data
          sed -i"*" "s|__HTTP_PROXY__|#{HTTP_PROXY}|g" /tmp/vagrantfile-user-data
          sed -i"*" "s|__HTTPS_PROXY__|#{HTTPS_PROXY}|g" /tmp/vagrantfile-user-data
          sed -i"*" "s|__NO_PROXY__|#{NO_PROXY}|g" /tmp/vagrantfile-user-data
          EOF
        end
        kHost.vm.provision :shell, :privileged => true,
        inline: <<-EOF
          sed -i"*" "/__PROXY_LINE__/d" /tmp/vagrantfile-user-data
          sed -i"*" "s,__DOCKER_OPTIONS__,#{DOCKER_OPTIONS},g" /tmp/vagrantfile-user-data
          sed -i"*" "s,__NOMAD_VERSION__,#{NOMAD_VERSION},g" /tmp/vagrantfile-user-data
          sed -i"*" "s,__CONSUL_VERSION__,#{CONSUL_VERSION},g" /tmp/vagrantfile-user-data
          sed -i"*" "s,__CHANNEL__,v#{CHANNEL},g" /tmp/vagrantfile-user-data
          sed -i"*" "s,__NAME__,#{hostname},g" /tmp/vagrantfile-user-data
          sed -i"*" "s|__MASTER_IP__|#{MASTER_IP}|g" /tmp/vagrantfile-user-data
          sed -i"*" "s|__DNS_DOMAIN__|#{DNS_DOMAIN}|g" /tmp/vagrantfile-user-data
          sed -i"*" "s|__ETCD_SEED_CLUSTER__|#{ETCD_SEED_CLUSTER}|g" /tmp/vagrantfile-user-data
          mv /tmp/vagrantfile-user-data /var/lib/coreos-vagrant/
        EOF
      end
    end
  end
end
