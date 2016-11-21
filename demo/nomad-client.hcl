# log verbosity level
log_level = "DEBUG"

# setup data dir
data_dir = "/tmp/nomad"

# this is client mode            
client {
    enabled = true
    servers = ["__MASTER_IP__:4647"]
}
