package v1

//OsCommand is for operating system related commands for block devices
type OsCommand struct {
	Command string
	Flag    string
}

//BlockDeviceInfo exposes the json output of lsblk:"blockdevices"
type BlockDeviceInfo struct {
	Blockdevices []Blockdevice `json:"blockdevices"`
}

//Blockdevices has block disk fields
type Blockdevice struct {
	Name       string        `json:"name"`               //block device name
	Majmin     string        `json:"maj:min"`            //major and minor block device number
	Rm         string        `json:"rm"`                 //is device removable
	Size       string        `json:"size"`               //size of device
	Ro         string        `json:"ro"`                 //is device read-only
	Type       string        `json:"type"`               //is device disk or partition
	Mountpoint string        `json:"mountpoint"`         //block device mountpoint
	Children   []Blockdevice `json:"children,omitempty"` //Blockdevice ...
}
