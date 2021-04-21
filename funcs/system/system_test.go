package system

import (
	"log"
	"testing"

	hostutil "github.com/shirou/gopsutil/host"
)

func TestSystem(t *testing.T) {
	info, _ := hostutil.Info()
	log.Printf("%s-%s-%s, %s, %s", info.Platform, info.PlatformFamily, info.PlatformVersion, info.KernelArch, info.KernelVersion)

}
