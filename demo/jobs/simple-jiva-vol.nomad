job "simple-vol" {
	datacenters = ["dc1"]

	# All tasks in this job must run on linux.
  constraint {
    attribute = "${attr.kernel.name}"
    value     = "linux"
  }
		
	# Configure the job to do rolling updates
	update {
		# Stagger updates every 10 seconds
		stagger = "10s"

		# Update a single task at a time
		max_parallel = 1
	}

	group "openebs" {
	
	  # All groups in this job should be scheduled on different hosts.
    constraint {
      operator  = "distinct_hosts"
      value     = "true"
    }

		restart {			
			attempts = 10
			interval = "5m"
			
			delay = "25s"
			mode = "delay"
		}

		task "jiva" {
			driver = "docker"

			config {
				image = "openebs/jiva"
				privileged = true
				command = "launch-simple-jiva"
				args = [ "simple-vol", "1g", "gotgt" ]
			}

			service {
				name = "${TASKGROUP}-jiva"
				tags = ["global", "openebs", "simple-vol"]
				port = "iscsi"
				check {
					name = "alive"
					type = "tcp"
					interval = "10s"
					timeout = "2s"
				}
			}

			resources {
				cpu = 500 # 500 MHz
				memory = 256 # 256MB
				network {
					mbits = 20
					port "iscsi" {
						static = "3260"
					}
				}
			}

		}
	}
}
