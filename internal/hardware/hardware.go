package hardware

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

func GetSystemSection() (string, error) {
	runTimeOS := runtime.GOOS

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return "", err
	}
	hostStat, err := host.Info()
	if err != nil {
		return "", err
	}

	output :=
		fmt.Sprintf("Hostname: %s Total memory: %d Used Memory: %d OS: %s",
			hostStat.Hostname, vmStat.Total, vmStat.Used, runTimeOS)

	return output, nil
}

func GetCPUSection() (string, error) {
	cpuStat, err := cpu.Info()
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("CPU: %s Cores: %d", cpuStat[0].ModelName, len(cpuStat))

	return output, nil
}

func GetDiskSection() (string, error) {
	diskStat, err := disk.Usage("/")
	if err != nil {
		return "", nil
	}

	output := fmt.Sprintf("Total disk space: %d Free disk space: %d", diskStat.Total, diskStat.Free)

	return output, nil
}
