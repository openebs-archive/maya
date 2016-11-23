# Nomad setup

This is an adaptation of [this link](http://rick-hightower.blogspot.in/2016/02/setting-up-nomad-and-consul-in-ec2.html)

However, this is a much more programmable version of above blog.

## Environment variables

You can set your custom environment values while calling vagrant up !!

```bash
# e.g. If you want to run 1 Server & 3 Clients then run below:
C_NODES=3 vagrant up

# e.g. If you want to run 3 Servers & 3 Clients then run below:
S_NODES=3 C_NODES=3 vagrant up
```

The envrionment variables that are currently available:

#### Installer versions
NOMAD_VERSION = ENV['NOMAD_VERSION'] || "0.5.0"
CONSUL_VERSION = ENV['CONSUL_VERSION'] || "0.7.1"

#### Server Nodes
S_NODES = ENV['S_NODES'] || 1

#### Client Nodes
C_NODES = ENV['C_NODES'] || 2

#### Server Memory & CPUs
S_MEM = ENV['S_MEM'] || 512
S_CPUS = ENV['S_CPUS'] || 1

#### Client Memory & CPUs
C_MEM = ENV['C_MEM'] || 1024
C_CPUS = ENV['C_CPUS'] || 1

#### Private IP address of server(s) & client(s)
BASE_SIP_ADDR = ENV['BASE_SIP_ADDR'] || "10.21.0"
BASE_CIP_ADDR = ENV['BASE_CIP_ADDR'] || "10.31.0"
