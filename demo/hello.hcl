job "helloworld-v1" {
  datacenters = ["dc1"]
  type = "service"

  update {
    stagger = "30s"
    max_parallel = 1
  }

  group "hello-group" {
    count = 1
    task "hello-task" {
      driver = "docker"
      config {
        image = "eveld/helloworld:1.0.0"
        port_map {
          http = 8080
        }
      }
      resources {
        cpu = 100
        memory = 200
        network {
          mbits = 1
          port "http" {}
        }
      }
    }
  }
}
