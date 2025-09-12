package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: SimpleContainer <os> <command>")
		return
	}

	switch os.Args[1] {

	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		memFlag := runCmd.String("memory", "", "Memory limit ('100M', '1G')")
		cpuFlag := runCmd.String("cpu", "", "CPU quota ('50000' for 50% of a core)")

		runCmd.Parse(os.Args[2:])

		if runCmd.NArg() < 2 {
			fmt.Println("Usage: ./simplecontainer run [--memory <limit>] [--cpu <quota>] <os> <command>")
			return
		}

		run(*memFlag, *cpuFlag, runCmd.Args())

	case "fork":
		fork()
	default:
		fmt.Println("Usage: SimpleContainer <os> <command>")
	}

}

func run(memLimit string, cpuQuota string, containerArgs []string) {

	printpid()

	// osName := os.Args[1]
	args := append([]string{"fork"}, containerArgs...)

	cmd := exec.Command("/proc/self/exe", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:  syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
		GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
	}

	setupCgroups()
	defer cleanupCgroups()

	setResourceLimits(memLimit, cpuQuota)

	must(cmd.Start())
	fmt.Printf("started process with PID: %d\n", cmd.Process.Pid)

	addProcessToCgroup(cmd.Process.Pid)

	must(cmd.Wait())
}

func fork() {

	printpid()
	fmt.Println(os.Args)
	osName := os.Args[2]
	command := os.Args[3]

	newRoot := filepath.Join("/home/vagrant/rootfs", osName)

	//by chroot
	// must(syscall.Chroot(newRoot))
	// must(syscall.Chdir("/"))

	must(syscall.Mount("proc", filepath.Join(newRoot, "proc"), "proc", syscall.MS_NOEXEC|syscall.MS_NODEV|syscall.MS_NOSUID, ""))
	must(syscall.Mount("tmpfs", filepath.Join(newRoot, "dev"), "tmpfs", syscall.MS_NOSUID, "mode=755,size=65536k")) //syscall.MS_IPV6_RTHDR_STRICT
	must(syscall.Mount("tmpfs", filepath.Join(newRoot, "tmp"), "tmpfs", syscall.MS_NOSUID, "mode=755,size=65536k"))

	//pivot root
	must(syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""))

	putOld := filepath.Join(newRoot, "put_old")

	must(os.MkdirAll(putOld, 0755))
	must(syscall.PivotRoot(newRoot, putOld))
	must(syscall.Chdir("/"))

	putOld = filepath.Base(putOld)
	must(syscall.Unmount(putOld, syscall.MNT_DETACH))
	must(syscall.Rmdir(putOld))

	cmd := exec.Command(command, os.Args[4:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	must(err)

	must(syscall.Unmount("proc", syscall.MNT_DETACH))
	must(syscall.Unmount("dev", syscall.MNT_DETACH))
	must(syscall.Unmount("tmp", syscall.MNT_DETACH))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func printpid() {
	fmt.Printf("PID: %d\n", os.Getpid())
}
