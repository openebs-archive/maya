region = "BANG-EAST"
datacenter = "dc2"
name = "my-vsm"
data_dir = "/tmp/mayaserver"
log_level = "ERR"
bind_addr = "192.168.0.1"
enable_debug = true
ports {
	http = 1234
}
addresses {
	http = "127.0.0.1"
}
advertise {
}
leave_on_interrupt = true
leave_on_terminate = true
enable_syslog = true
syslog_facility = "LOCAL1"
http_api_response_headers {
	Access-Control-Allow-Origin = "*"
}
