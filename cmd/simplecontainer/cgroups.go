package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const cgroupPath = "/sys/fs/cgroup/mycontainer"

func setupCgroups() {
	must(os.MkdirAll(cgroupPath, 0755))
}

func setResourceLimits(memLimit, cpuQuota string) {
	if memLimit != "" {
		memInBytes, err := parseMemory(memLimit)
		must(err)
		must(os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte(strconv.FormatInt(memInBytes, 10)), 0700))
	}

	if cpuQuota != "" {
		limit := fmt.Sprintf("%s 100000", cpuQuota)
		must(os.WriteFile(filepath.Join(cgroupPath, "cpu.max"), []byte(limit), 0700))
	}
}

func addProcessToCgroup(pid int) {
	must(os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700))
}

func cleanupCgroups() {

	if err := os.RemoveAll(cgroupPath); err != nil {
		fmt.Printf("Warning: failed to remove cgroup path %s: %v\n", cgroupPath, err)
	}

}

func parseMemory(memStr string) (int64, error) {
	memStr = strings.ToUpper(memStr)
	var multiplier int64 = 1

	if strings.HasSuffix(memStr, "M") {
		multiplier = 1024 * 1024
		memStr = strings.TrimSuffix(memStr, "M")
	} else if strings.HasSuffix(memStr, "G") {
		multiplier = 1024 * 1024 * 1024
		memStr = strings.TrimSuffix(memStr, "G")
	}

	val, err := strconv.ParseInt(memStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return val * multiplier, nil
}
