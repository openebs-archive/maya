log_level = "DEBUG"
data_dir = "/tmp/nomad"


client {
  enabled = true
  options = {
    "docker.privileged.enabled" = "true"
    "driver.raw_exec.enable" = "1"
  }
}
