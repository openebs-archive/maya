package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/mitchellh/cli"
)

//Installer is a interface for Install
type Installer interface {
	Install() int
}

//InstallCommand is a collection
type InstallCommand struct {
	// To control this CLI's display
	UI cli.Ui

	// OS command to execute
	Cmd *exec.Cmd
}

//MayaAsNomadInstaller is a collection
type MayaAsNomadInstaller struct {
	InstallCommand

	// if the installation is in server mode or client mode
	isMaster bool

	// all maya master ips, in a comma separated format
	masterIps string

	// all maya openebs ips, in a comma separated format
	clientIps string

	// self ip address
	selfIP string

	// the provided master ips in a format understood by Nomad & Consul
	fmtMasterIps string

	// the provided master ips with rpc ports in a format understood by Nomad
	fmtMasterIpnports string

	// a trimmed version of selfIP
	selfIPTrim string

	// formatted etcd initial cluster
	etcdCluster string

	//contains version info
	nomad  string
	consul string
}

// Install is the public command
func (c *MayaAsNomadInstaller) Install() (runop int) {

	if runop = c.verifyBootstrap(); runop != 0 {
		//Run the bootstrap only if required
		if runop = c.bootstrap(); runop != 0 {
			return
		}
	}

	if runop = c.initAsClient(); runop != 0 {
		return
	}

	if runop = c.installDocker(); runop != 0 {
		return
	}

	if runop = c.installConsul(); runop != 0 {
		return
	}

	if runop = c.setConsulAsClient(); runop != 0 {
		return
	}

	if runop = c.startConsulAsClient(); runop != 0 {
		return
	}

	if runop = c.installNomad(); runop != 0 {
		return
	}

	if runop = c.setNomadAsClient(); runop != 0 {
		return
	}

	if runop = c.startNomadAsClient(); runop != 0 {
		return
	}
	//Not installing the etcd on Maya-host
	//if runop = c.installEtcd(); runop != 0 {
	//	return
	//}

	//if runop = c.setEtcd(); runop != 0 {
	//	return
	//}

	//if runop = c.startEtcd(); runop != 0 {
	//	return
	//}

	return
}

func (c *MayaAsNomadInstaller) bootstrap() (runop int) {

	c.Cmd = exec.Command("curl", "-sSL", BootstrapScriptPath, "-o", BootstrapScript)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error(fmt.Sprintf("Install failed: Bootstrap failed: Could not fetch file: %s", BootstrapScriptPath))

		c.Cmd = exec.Command("rm", "-rf", BootstrapScript)
		execute(c.Cmd, c.UI)

		return
	}

	c.Cmd = exec.Command("sh", "./"+BootstrapScript)
	runop = execute(c.Cmd, c.UI)

	c.Cmd = exec.Command("rm", "-rf", BootstrapScript)
	execute(c.Cmd, c.UI)

	if runop != 0 {
		c.UI.Error("Install failed: Error while bootstraping")
	}

	return
}

func (c *MayaAsNomadInstaller) verifyBootstrap() (runop int) {

	c.Cmd = exec.Command("ls", MayaScriptsPath)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error(fmt.Sprintf("Install failed: Bootstrap failed: Missing path: %s", MayaScriptsPath))
	}

	return
}

//TODO
func (c *MayaAsNomadInstaller) validateIPs() (runop int) {
	return 0
}

// Set the instance variables i.e. properties of
// MayaAsNomadInstaller
func (c *MayaAsNomadInstaller) initAsClient() (runop int) {

	var masterIparr []string
	var clientIParr []string
	var ipTrimmed string

	if len(strings.TrimSpace(c.selfIP)) == 0 {
		c.Cmd = exec.Command("sh", GetPrivateIPScript)

		if runop = execute(c.Cmd, c.UI, &c.selfIP); runop != 0 {
			c.UI.Error("Install failed: Error fetching local IP address")
			return
		}
	}

	if len(strings.TrimSpace(c.selfIP)) == 0 {
		c.UI.Error("Install failed: IP address could not be determined")
		return 1
	}

	// Stuff with client ips
	c.selfIPTrim = strings.Replace(c.selfIP, ".", "", -1)

	if len(strings.TrimSpace(c.clientIps)) > 0 {
		clientIParr = strings.Split(strings.TrimSpace(c.clientIps), ",")
	}

	clientIParr = append(clientIParr, c.selfIP)

	for _, clientIP := range clientIParr {
		clientIP = strings.TrimSpace(clientIP)

		if len(clientIP) == 0 {
			continue
		}

		ipTrimmed = strings.Replace(clientIP, ".", "", -1)

		if len(c.etcdCluster) > 0 {
			c.etcdCluster = c.etcdCluster + ","
		}

		c.etcdCluster = c.etcdCluster + ipTrimmed + "=https://" + clientIP + ":2380"

	}

	// Stuff with master ips
	if len(strings.TrimSpace(c.masterIps)) > 0 {
		masterIparr = strings.Split(strings.TrimSpace(c.masterIps), ",")
	}

	for _, masterIP := range masterIparr {

		masterIP = strings.TrimSpace(masterIP)

		if len(masterIP) == 0 {
			continue
		}

		if len(c.fmtMasterIps) > 0 {
			c.fmtMasterIps = c.fmtMasterIps + ","
		}

		if len(c.fmtMasterIpnports) > 0 {
			c.fmtMasterIpnports = c.fmtMasterIpnports + ","
		}

		c.fmtMasterIps = c.fmtMasterIps + `"` + masterIP + `"`
		c.fmtMasterIpnports = c.fmtMasterIpnports + `"` + masterIP + `:4647"`
	}

	return
}

func (c *MayaAsNomadInstaller) installDocker() (runop int) {

	c.Cmd = exec.Command("bash", InstallDockerScript)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Error installing docker")
	}

	return
}

//func (c *MayaAsNomadInstaller) installEtcd() int {

//	var runop int

//	c.Cmd = exec.Command("sh", InstallEtcdScript)

//	if runop = execute(c.Cmd, c.UI); runop != 0 {
//		c.UI.Error("Install failed: Error installing etcd")
//	}

//	return runop
//}

func (c *MayaAsNomadInstaller) installConsul() (runop int) {

	c.Cmd = exec.Command("sh", InstallConsulScript, c.consul)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Error installing consul")
	}

	return
}

func (c *MayaAsNomadInstaller) installNomad() (runop int) {

	c.Cmd = exec.Command("sh", InstallNomadScript, c.nomad)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Error installing nomad")
	}

	return
}

//func (c *MayaAsNomadInstaller) startEtcd() int {

//	var runop int

//	c.Cmd = exec.Command("sh", StartEtcdScript)

//	if runop := execute(c.Cmd, c.UI); runop != 0 {
//		c.UI.Error("Install failed: Systemd failed: Error starting etcd")
//	}

//	return runop
//}

func (c *MayaAsNomadInstaller) startConsulAsClient() (runop int) {

	c.Cmd = exec.Command("sh", StartConsulClientScript)

	if runop := execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Systemd failed: Error starting consul in client mode")
	}

	return
}

func (c *MayaAsNomadInstaller) startNomadAsClient() (runop int) {

	c.Cmd = exec.Command("sh", StartNomadClientScript)

	if runop := execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Systemd failed: Error starting nomad in client mode")
	}

	return
}

//func (c *MayaAsNomadInstaller) setEtcd() int {

//	var runop int

//	c.Cmd = exec.Command("sh", SetEtcdScript, c.selfIP, c.selfIPTrim, c.etcdCluster)

//	if runop = execute(c.Cmd, c.UI); runop != 0 {
//		c.UI.Error("Install failed: Error setting etcd")
//	}

//	return runop
//}

func (c *MayaAsNomadInstaller) setConsulAsClient() (runop int) {

	c.Cmd = exec.Command("sh", SetConsulAsClientScript, c.selfIP, c.fmtMasterIps)

	c.Cmd = exec.Command("sh", SetConsulAsClientScript, c.selfIP, c.fmtMasterIps)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Error setting consul as client")
	}

	return
}

func (c *MayaAsNomadInstaller) setNomadAsClient() (runop int) {

	c.Cmd = exec.Command("sh", SetNomadAsClientScript, c.selfIP, c.fmtMasterIpnports)

	c.Cmd = exec.Command("sh", SetNomadAsClientScript, c.selfIP, c.fmtMasterIpnports)

	if runop = execute(c.Cmd, c.UI); runop != 0 {
		c.UI.Error("Install failed: Error setting nomad as client")
	}

	return
}
