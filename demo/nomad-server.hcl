# log verbosity level
log_level = "DEBUG"

# setup data directory
data_dir = "/tmp/nomad"

bind_addr = "__THE_IPV4__"

region = "hsr"

datacenter = "mayadc"

# this lets the server gracefully leave after a SIGTERM
leave_on_terminate = true

# We need to specify our host's IP because we can't
# advertise 0.0.0.0 to other nodes in the cluster.
advertise {
  http = "__THE_IPV4__:4646"
  rpc = "__THE_IPV4__:4647"
  serf = "__THE_IPV4__:4648"
}

server {
    enabled = true
    bootstrap_expect = 1
}

# allow Nomad to automatically find its peers through Consul
consul {
  server_service_name = "nomad"
  server_auto_join = true
  client_service_name = "nomad-client"
  client_auto_join = true
  auto_advertise = true
}
