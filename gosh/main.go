package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

var running bool = true

func main() {
	fmt.Printf("gosh\n")
	// Set up non-canonical mode for reading from /dev/console
	setNonCanonicalMode()

	for running {
		fmt.Print("\033[31m#\033[39m ")
		cmdString, err := readCommand()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		err = runCommand(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func setNonCanonicalMode() {
	// Retrieve current terminal settings
	var termios unix.Termios
	_, _, err := unix.Syscall(unix.SYS_IOCTL, os.Stdin.Fd(), unix.TCGETS, uintptr(unsafe.Pointer(&termios)))
	if err != 0 {
		fmt.Println("Error getting terminal settings:", err)
		os.Exit(1)
	}

	// Set non-canonical mode
	termios.Lflag &^= unix.ICANON | unix.ECHO
	_, _, err = unix.Syscall(unix.SYS_IOCTL, os.Stdin.Fd(), unix.TCSETS, uintptr(unsafe.Pointer(&termios)))
	if err != 0 {
		fmt.Println("Error setting terminal settings:", err)
		os.Exit(1)
	}
}

func readCommand() (string, error) {
	var buf [1]byte
	var cmdString strings.Builder
	for {
		n, err := unix.Read(int(os.Stdin.Fd()), buf[:])
		if err != nil {
			return "", err
		}
		if n > 0 {
			char := buf[0]
			if char == '\n' {
				fmt.Println()
				return cmdString.String(), nil
			}
			fmt.Print(string(char))
			cmdString.WriteByte(char)
		}
	}
}

func runCommand(commandStr string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	if len(arrCommandStr) == 0 {
		// Empty command, just return
		return nil
	}
	switch arrCommandStr[0] {
	case "exit":
		running = false
		// add another case here for custom commands.
	case "help":
		fmt.Printf("golinux/gosh commands:\n\nhelp - shows this menu\nexit - exits (also kernel panics)\nreboot - reboots the system\nshutdown - shuts down\ntest - tests functionality\ntasklist - lists processes\ngr - goroutine test\n")
	case "reboot":
		fmt.Printf("syncing disks...\n")
		unix.Sync()
		fmt.Printf("rebooting...\n")
		unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART)
	case "shutdown":
		fmt.Printf("syncing disks...\n")
		unix.Sync()
		fmt.Printf("shutting down...\n")
		unix.Reboot(unix.LINUX_REBOOT_CMD_POWER_OFF)
	case "test":
		fmt.Printf("Starting tests...\n")
		//testing()
	case "tasklist":
		matches, _ := filepath.Glob("/proc/*/exe")
		for _, file := range matches {
			target, _ := os.Readlink(file)
			if len(target) > 0 {
				fmt.Printf("%+v\n", target)
			}
		}
	case "gr":
		go test1()
		go test2()
	default:
		fmt.Printf("invalid command\n")
	}
	return nil
}

func test1() {
	fmt.Printf("%v\n", 100+43)
	return
}

func test2() {
	fmt.Printf("%v\n", 40+2)
	return
}
