package api

import (
	"github.com/zcalusic/sysinfo"
)

type Distro struct {
	name string
	version string
}

func GetCurrentDistro() Distro {
	var si sysinfo.SysInfo
	si.GetSysInfo()

	return Distro{si.OS.Name, si.OS.Release}
}
