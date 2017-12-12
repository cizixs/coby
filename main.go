package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
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

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func initProcess() {
	// initial setup, before running the real command
	mountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(mountFlags), ""); err != nil {
		fmt.Printf("mount proc failed: %v", err)
		return
	}

	if err := syscall.Sethostname([]byte("container")); err != nil {
		fmt.Printf("set hostname failed: %v", err)
	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("run command in container failed: %v", err)
		os.Exit(1)
	}

	// umount proc to make host work normally
	if err := syscall.Unmount("/proc", 0); err != nil {
		fmt.Printf("umount proc failed: %v", err)
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
