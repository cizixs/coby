package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const (
	cgMemoryPath = "/sys/fs/cgroup/memory"
)

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"init"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET,
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	//setup cgroup limit for container
	// 1. create cgroup memeory subssytem for container
	fmt.Printf("coby init run in pid: %d\n", cmd.Process.Pid)
	if err := os.MkdirAll(path.Join(cgMemoryPath, "coby", "container"), 0755); err != nil {
		fmt.Printf("failed to setup memeory cgroup: %v", err)
		os.Exit(1)
	}

	// 2. limit memeory usage
	if err := ioutil.WriteFile(path.Join(cgMemoryPath, "coby", "container", "memory.limit_in_bytes"), []byte("100m"), 0644); err != nil {
		fmt.Printf("failed to limit memeory usage in cgroup: %v", err)
		os.Exit(1)
	}

	// 3. add container to cgroup
	if err := ioutil.WriteFile(path.Join(cgMemoryPath, "coby", "container", "tasks"), []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		fmt.Printf("failed to add container to memeory cgroup: %v", err)
		os.Exit(1)
	}

	defer func() {
		os.RemoveAll(path.Join(cgMemoryPath, "coby", "container"))
	}()

	cmd.Wait()
}

func initProcess() {
	// initial setup, before running the real command

	// use chroot to jail root file system
	if err := syscall.Chroot("./rootfs"); err != nil {
		fmt.Printf("Chroot failed: %v", err)
		os.Exit(1)
	}

	if err := os.Chdir("/"); err != nil {
		fmt.Printf("Change to root path failed: %v", err)
	}

	// setup mount points
	mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), ""); err != nil {
		fmt.Printf("mount proc failed: %v", err)
		return
	}

	if err := syscall.Mount("tmp", "/tmp", "tmpfs", 0, ""); err != nil {
		fmt.Printf("mount /tmp directory failed: %v", err)
		os.Exit(1)
	}

	// setup hostname
	// FIXME(cizixs): use container name or auto-generate hostname
	if err := syscall.Sethostname([]byte("container")); err != nil {
		fmt.Printf("set hostname failed: %v", err)
	}

	// Now, it's time to run user-specified command in container
	name, err := exec.LookPath(os.Args[2])
	if err != nil {
		fmt.Printf("run command in container failed: %v\n", err)
		os.Exit(1)
	}

	if err := syscall.Exec(name, os.Args[2:], os.Environ()); err != nil {
		fmt.Printf("run command in container failed: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s run\n", os.Args[0])
		return
	}

	switch os.Args[1] {
	case "run":
		run()
	case "init":
		initProcess()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
	}
}
