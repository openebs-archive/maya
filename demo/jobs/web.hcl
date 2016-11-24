job "web" {
  datacenters = ["dc1"]

  type = "service"

  constraint {
    distinct_hosts = true
  }

  update {
    stagger = "30s"
    max_parallel = 1
  }

  group "servers" {
    count = 1

    task "nginx" {
      driver = "docker"

      config {
        image = "nginx"
        port_map = {
          http = 80
        }
      }

      service {
        name = "web"
        tags = ["lb-external"]
        port = "http"
        check {
          type = "http"
          path = "/"
          interval = "10s"
          timeout = "4s"
        }
      }

      resources {
        cpu = 500
        memory = 256
        network {
          mbits = 10
          port "http" {
            static = 11080
          }
        }
      }
    }
  }

}
