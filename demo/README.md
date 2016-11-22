# nomad-vagrant-coreos-cluster
Turnkey **[Nomad](https://github.com/hashicorp/nomad)**
cluster with **[Consul](https://github.com/hashicorp/consul)** integration
on top of **[Vagrant](https://www.vagrantup.com)** (1.7.2+) and
**[CoreOS](https://coreos.com)**.

## Pre-requisites

 * **[Vagrant](https://www.vagrantup.com)**
 * a supported Vagrant hypervisor
 	* **[Virtualbox](https://www.virtualbox.org)** (the default)
 	* **[Parallels Desktop](http://www.parallels.com/eu/products/desktop/)**
 	* **[VMware Fusion](http://www.vmware.com/products/fusion)** or **[VMware Workstation](http://www.vmware.com/products/workstation)**

### MacOS X

On **MacOS X** (and assuming you have [homebrew](http://brew.sh) already installed) run

```
brew update
brew install wget
```

## Deploy Nomad

Current `Vagrantfile` will bootstrap one VM with everything needed to become a Nomad _server_ and, by default, a couple VMs with everything needed to become Nomad clients.
You can change the number of minions and/or the Nomad version by setting environment variables **NODES** and **NOMAD_VERSION**, respectively. [You can find more details below](#customization).

```
vagrant up
```

### Linux or MacOS host

Nomad cluster is ready. but you need to set-up some environment variables that we have already provisioned for you. In the current terminal window, run:

```
source ~/.bash_profile
```

New terminal windows will have this set for you.

## Run example

Just for the sake of curiosity, the cluster status should be something like:
```
$ nomad server-members
Name           Address       Port  Status  Leader  Protocol  Build     Datacenter  Region
server.global  172.17.9.100  4648  alive   true    2         0.4.1     dc1         global

$ nomad node-status
ID        DC   Name       Class   Drain  Status
363a2321  dc1  client-02  <none>  false  ready
fa511f91  dc1  client-01  <none>  false  ready
```

Now, let's deploy Nginx example:

```
$ nomad run web.hcl
==> Monitoring evaluation "bb78252e"
    Evaluation triggered by job "web"
    Allocation "8a0ed733" created: node "363a2321", group "servers"
    Evaluation status changed: "pending" -> "complete"
==> Evaluation "bb78252e" finished with status "complete"
```

Check its status:

```
$ nomad status web
ID          = web
Name        = web
Type        = service
Priority    = 50
Datacenters = dc1
Status      = running
Periodic    = false

Allocations
ID        Eval ID   Node ID   Task Group  Desired  Status
8a0ed733  bb78252e  363a2321  servers     run      running
```

and its registration on Consul:

```
$ curl "${NOMAD_ADDR%:*}:8500/v1/catalog/service/web"
[
  {
    "Node":"client-01",
    "Address":"172.17.9.101",
    "ServiceID":"nomad-registered-service-22ba6909-dacf-6965-519c-9cd7f8a67844",
    "ServiceName":"web",
    "ServiceTags":[
      "lb-external"
    ],
    "ServiceAddress":"172.17.9.101",
    "ServicePort":11080,
    "ServiceEnableTagOverride":false,
    "CreateIndex":102,
    "ModifyIndex":104
  }
]
```

## Clean-up

```
vagrant destroy -f
```

If you've set `NODES` or any other variable when deploying, please make sure you set it in `vagrant destroy` call above, like:

```
NODES=3 vagrant destroy -f
```

## Notes about hypervisors

### Virtualbox

**VirtualBox** is the default hypervisor, and you'll probably need to disable its DHCP server
```
VBoxManage dhcpserver remove --netname HostInterfaceNetworking-vboxnet0
```

### Parallels

If you are using **Parallels Desktop**, you need to install **[vagrant-parallels](http://parallels.github.io/vagrant-parallels/docs/)** provider 
```
vagrant plugin install vagrant-parallels
```
Then just add `--provider parallels` to the `vagrant up` invocations above.

### VMware
If you are using one of the **VMware** hypervisors you must **[buy](http://www.vagrantup.com/vmware)** the matching  provider and, depending on your case, just add either `--provider vmware-fusion` or `--provider vmware-workstation` to the `vagrant up` invocations above.

## Private Docker Repositories

If you want to use Docker private repositories look for **DOCKERCFG** bellow.

## Customization
### Environment variables
Most aspects of your cluster setup can be customized with environment variables. Right now the available ones are:

 - **NODES** sets the number of nodes (minions).

   Defaults to **2**.
 - **CHANNEL** sets the default CoreOS channel to be used in the VMs.

   Defaults to **alpha**.

   While by convenience, we allow an user to optionally consume CoreOS' *beta* or *stable* channels please do note that as both Nomad and CoreOS are quickly evolving platforms we only expect our setup to behave reliably on top of CoreOS _alpha_ channel.
   So, **before submitting a bug**, either in [this](https://github.com/pires/nomad-vagrant-coreos-cluster/issues) project, or in ([Nomad](https://github.com/hashicorp/nomad/issues) or [CoreOS](https://github.com/coreos/bugs/issues)) **make sure it** (also) **happens in the** (default) **_alpha_ channel** :smile:
 - **COREOS_VERSION** will set the specific CoreOS release (from the given channel) to be used.

   Default is to use whatever is the **latest** one from the given channel.
 - **SERIAL_LOGGING** if set to *true* will allow logging from the VMs serial console.

   Defaults to **false**. Only use this if you *really* know what you are doing.
 - **SERVER_MEM** sets the server VM memory.

   Defaults to **512** (in MB)
 - **SERVER_CPUS** sets the number of vCPUs to be used by the server VM.

   Defaults to **1**.
 - **CLIENT_MEM** sets the client VMs memory.

   Defaults to **2048** (in MB)
 - **CLIENT_CPUS** sets the number os vCPUs to be used by the client VMs.

   Defaults to **1**.
 - **DOCKERCFG** sets the location of your private docker repositories (and keys) configuration. However, this is only usable if you set **USE_DOCKERCFG=true**.

   Defaults to "**~/.dockercfg**".

   You can create/update a *~/.dockercfg* file at any time
   by running `docker login <registry>.<domain>`. All nodes will get it automatically,
   at `vagrant up`, given any modification or update to that file.

 - **DOCKER_OPTIONS** sets the additional `DOCKER_OPTS` for docker service on both master and the nodes. Useful for adding params such as `--insecure-registry`.

 - **NOMAD_VERSION** defines the specific Nomad version being used.

   Defaults to `0.4.1`.

So, in order to start, say, a Nomad cluster with 3 client nodes, 4GB of RAM and 2 vCPUs per client, one would run:

```
CLIENT_MEM=4096 CLIENT_CPUS=2 NODES=3 vagrant up
```

**Please do note** that if you were using non default settings to startup your
cluster you *must* also use those exact settings when invoking
`vagrant {up,ssh,status,destroy}` to communicate with any of the nodes in the cluster as otherwise
things may not behave as you'd expect.

### Synced Folders
You can automatically mount in your *guest* VMs, at startup, an arbitrary
number of local folders in your host machine by populating accordingly the
`synced_folders.yaml` file in your `Vagrantfile` directory. For each folder
you which to mount the allowed syntax is...

```yaml
# the 'id' of this mount point. needs to be unique.
- name: foobar
# the host source directory to share with the guest(s).
  source: /foo
# the path to mount ${source} above on guest(s)
  destination: /bar
# the mount type. only NFS makes sense as, presently, we are not shipping
# hypervisor specific guest tools. defaults to `true`.
  nfs: true
# additional options to pass to the mount command on the guest(s)
# if not set the Vagrant NFS defaults will be used.
  mount_options: 'nolock,vers=3,udp,noatime'
# if the mount is enabled or disabled by default. default is `true`.
  disabled: false
```

## Licensing

This work is [open-source](http://opensource.org/osd), and is licensed under the [Apache License, Version 2.0](http://opensource.org/licenses/Apache-2.0).


