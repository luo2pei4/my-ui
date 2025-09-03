package command

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

type BlockDevice struct {
	Name       string        `json:"name,omitempty"`
	Type       string        `json:"type,omitempty"`
	Size       int64         `json:"size,omitempty"`
	Rota       bool          `json:"rota,omitempty"`
	Serial     string        `json:"serial,omitempty"`
	WWN        string        `json:"wwn,omitempty"`
	Vendor     string        `json:"vendor,omitempty"`
	Model      string        `json:"model,omitempty"`
	Rev        string        `json:"rev,omitempty"`
	MountPoint string        `json:"mountpoint,omitempty"`
	PartUUID   string        `json:"partuuid,omitempty"`
	UUID       string        `json:"uuid,omitempty"`
	PTUUID     string        `json:"ptuuid,omitempty"`
	FSAvail    string        `json:"fsavail,omitempty"`
	FSSize     string        `json:"fssize,omitempty"`
	FSUsed     string        `json:"fsused,omitempty"`
	FSType     string        `json:"fstype,omitempty"`
	Children   []BlockDevice `json:"children,omitempty"`
}

const (
	DiskLabelBsd   = "bsd"
	DiskLabelLoop  = "loop"
	DiskLabelGpt   = "gpt"
	DiskLabelMac   = "mac"
	DiskLabelMsdos = "msdos"
	DiskLabelPc98  = "pc98"
	DiskLabelSun   = "sun"
)

const (
	Lsblk     = "lsblk --paths --json --bytes --fs --output NAME,TYPE,SIZE,ROTA,SERIAL,WWN,VENDOR,MODEL,REV,MOUNTPOINT,PARTUUID,UUID,PTUUID,FSAVAIL,FSSIZE,FSUSED,FSTYPE"
	outputKey = "blockdevices"
)

var (
	pureNumberReg = regexp.MustCompile(`[0-9]+`)
)

var ignoreDeviceType = map[string]struct{}{
	"loop": {},
	"rom":  {},
	"usb":  {},
}

func ignoreDevice(devType string) bool {
	_, ok := ignoreDeviceType[devType]
	return ok
}

// GetBlockDevices get block devices
func GetBlockDevices(result []byte) ([]BlockDevice, error) {
	rawOut := make(map[string][]BlockDevice, 1)
	err := json.Unmarshal(result, &rawOut)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal output to BlockDevice instance, error: %v", err)
	}
	var (
		devs []BlockDevice
		ok   bool
	)
	if devs, ok = rawOut[outputKey]; !ok {
		return nil, fmt.Errorf("unexpected lsblk output format, missing \"%s\" key", outputKey)
	}
	res := make([]BlockDevice, 0, len(devs))
	for _, d := range devs {
		if !ignoreDevice(d.Type) {
			res = append(res, d)
		}
	}
	return res, nil
}

func (dev *BlockDevice) IsRootDisk() bool {
	hasRoot := false
	for _, child := range dev.Children {
		if child.MountPoint == "/" {
			hasRoot = true
		}
		if len(child.Children) > 0 {
			hasRoot = child.IsRootDisk()
		}
	}
	return hasRoot
}

func (dev *BlockDevice) IsMounted() (bool, []string) {
	isMounted := false
	mps := make([]string, 0)
	for _, child := range dev.Children {
		if len(child.MountPoint) != 0 {
			isMounted = true
			mps = append(mps, child.MountPoint)
		}
		if len(child.Children) > 0 {
			mounted, mountpoints := child.IsMounted()
			isMounted = mounted
			mps = append(mps, mountpoints...)
		}
	}
	return isMounted, mps
}

// devMpMap：key: device；value: mountpoint
// mpDevMap：key: mountpoint；value: device
func (dev *BlockDevice) GetMountInfo() (devMpMap, mpDevMap map[string]string) {
	devMpMap = make(map[string]string)
	mpDevMap = make(map[string]string)
	return getMountInfo(dev, devMpMap, mpDevMap)
}

func getMountInfo(dev *BlockDevice, devMpMap, mpDevMap map[string]string) (map[string]string, map[string]string) {
	for _, child := range dev.Children {
		if len(child.MountPoint) != 0 {
			devMpMap[child.Name] = child.MountPoint
			mpDevMap[child.MountPoint] = child.Name
		}
		if len(child.Children) > 0 {
			devMpMap, mpDevMap = getMountInfo(&child, devMpMap, mpDevMap)
		}
	}
	return devMpMap, mpDevMap
}

func (dev *BlockDevice) UsedCapacity() int64 {
	var used int64
	for _, child := range dev.Children {
		if pureNumberReg.MatchString(child.FSUsed) {
			tmp, _ := strconv.ParseInt(child.FSUsed, 10, 64)
			used += tmp
		}
		if len(child.Children) > 0 {
			used += child.UsedCapacity()
		}
	}
	return used
}

func (dev *BlockDevice) GetParts() []BlockDevice {
	mps := make([]BlockDevice, 0)
	for _, child := range dev.Children {
		if child.Type == "part" {
			mps = append(mps, child)
		}
		if len(child.Children) > 0 {
			parts := child.GetParts()
			mps = append(mps, parts...)
		}
	}
	return mps
}
