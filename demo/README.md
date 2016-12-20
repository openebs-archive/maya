# Programmable Maya demo setup

This is an adaptation of [this link](http://rick-hightower.blogspot.in/2016/02/setting-up-nomad-and-consul-in-ec2.html)

You can create your custom demo environment while calling vagrant up !!

```bash
# e.g. If you want to run 1 Server & 3 Clients then run below:
C_NODES=3 vagrant up

# e.g. If you want to run 3 Servers & 3 Clients then run below:
S_NODES=3 C_NODES=3 vagrant up
```

### Pre-requisites
- Ubuntu 
- Vagrant
- (optional) Git to checkout the Vagrantfile 


The envrionment variables that are currently available:


#### Server Nodes

```bash
S_NODES = ENV['S_NODES'] || 1
```

#### Client Nodes

```bash
C_NODES = ENV['C_NODES'] || 2
```

#### Server Memory & CPUs

```bash
S_MEM = ENV['S_MEM'] || 512
S_CPUS = ENV['S_CPUS'] || 1
```

#### Client Memory & CPUs

```bash
C_MEM = ENV['C_MEM'] || 1024
C_CPUS = ENV['C_CPUS'] || 1
```

