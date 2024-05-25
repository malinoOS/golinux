package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var running bool = true
var oldState *syscall.Termios
var Version string = "undefined"

func main() {
	fmt.Printf("fallback shell v%v\n", Version)
	setNonCanonicalMode()
	defer resetTerminalMode()

	for running {
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Critical error while getting working directory:\n%v", err)
			return
		}
		fmt.Printf("\033[31m%v #\033[39m ", currentDir)
		cmdString := readCommand()
		err = runCommand(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func readCommand() string {
	var buf [1]byte
	var cmdString strings.Builder
	for {
		n, err := syscall.Read(int(os.Stdin.Fd()), buf[:])
		if err != nil {
			fmt.Printf("Critical error while reading characters:\n%v", err)
			return ""
		}
		if n > 0 {
			char := buf[0]
			if char == '\n' {
				fmt.Println()
				return cmdString.String()
			} else if char == 127 { // ASCII code for backspace
				if cmdString.Len() > 0 {
					cmd := cmdString.String()
					if len(cmd) > 1 {
						cmdString.Reset()
						cmdString.WriteString(cmd[:len(cmd)-1])
						fmt.Print("\b \b")
					} else {
						cmdString.Reset()
						fmt.Print("\b \b")
					}
				}
			} else {
				fmt.Print(string(char))
				cmdString.WriteByte(char)
			}
		}
	}
}

func runCommand(commandStr string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	if len(arrCommandStr) == 0 {
		return nil
	}
	switch arrCommandStr[0] {
	case "exit":
		running = false
	case "help":
		fmt.Println("golinux/gosh commands:\n\nhelp - shows this menu\nexit - exits (also kernel panics)\nreboot - reboots the system\nshutdown - shuts down\ntasklist - lists processes\ngr - goroutine test\ncd [folder] - change directory\nls - list the current directory\ncat [file] - dump file")
	case "reboot":
		fmt.Println("syncing disks...")
		syscall.Sync()
		fmt.Println("rebooting...")
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	case "shutdown":
		fmt.Println("syncing disks...")
		syscall.Sync()
		fmt.Println("shutting down...")
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	case "tasklist":
		matches, _ := filepath.Glob("/proc/*/exe")
		for _, file := range matches {
			pid := filepath.Base(filepath.Dir(file))
			target, _ := os.Readlink(file)
			if len(target) > 0 {
				fmt.Printf("PID: %s, Process: %+v\n", pid, target)
			}
		}
	case "gr":
		go test1()
		go test2()
	case "cd":
		cd(arrCommandStr)
	case "ls":
		currentDir, _ := os.Getwd()
		entries, err := os.ReadDir(currentDir)
		if err != nil {
			fmt.Printf("ls: could not list directory: %v\n", err)
			return nil
		}
		for _, e := range entries {
			if e.IsDir() {
				fmt.Print("\033[94m")
			} else {
				fmt.Print("\033[39m")
			}
			fmt.Printf("%v ", e.Name())
		}
		fmt.Print("\033[39m")
	case "cat":
		file := ""
		if strings.HasPrefix(arrCommandStr[1], "/") {
			file = arrCommandStr[1]
		} else {
			currentDir, _ := os.Getwd()
			file = fmt.Sprintf("%v/%v", currentDir, arrCommandStr[1])
		}
		dat, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("cat: could not read file: %v\n", err)
			return nil
		}
		fmt.Println(string(dat))
	default:
		fmt.Println("invalid command")
	}
	return nil
}

func cd(arrCommandStr []string) {
	dir := ""
	if arrCommandStr[1] == ".." {
		currentDir, _ := os.Getwd()
		dir = filepath.Dir(currentDir)
	} else if strings.HasPrefix(arrCommandStr[1], "/") {
		dir = arrCommandStr[1]
	} else {
		currentDir, _ := os.Getwd()
		dir = fmt.Sprintf("%v/%v", currentDir, arrCommandStr[1])
	}
	fileInfo, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("cd: could not change directory: %v doesn't exist\n", dir)
			return
		} else {
			fmt.Printf("cd: could not change directory: %v\n", err.Error())
			return
		}
	} else {
		if fileInfo.IsDir() {
			err := syscall.Chdir(dir)
			if err != nil {
				fmt.Printf("cd: could not change directory: %v\n", err.Error())
				return
			}
		} else {
			fmt.Printf("cd: could not change directory: %v isn't a directory\n", dir)
			return
		}
	}
}

func test1() {
	fmt.Printf("%v\n", 100+43)
	return
}

func test2() {
	fmt.Printf("%v\n", 40+2)
	return
}

func setNonCanonicalMode() {
	fd := int(os.Stdin.Fd())
	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	if errno != 0 {
		fmt.Printf("Error getting terminal attributes: %v\n", errno)
		os.Exit(1)
	}
	oldState = &termios
	termios.Lflag &^= syscall.ICANON | syscall.ECHO
	termios.Cc[syscall.VMIN] = 1
	termios.Cc[syscall.VTIME] = 0
	_, _, errno = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	if errno != 0 {
		fmt.Printf("Error setting terminal attributes: %v\n", errno)
		os.Exit(1)
	}
}

func resetTerminalMode() {
	if oldState != nil {
		fd := int(os.Stdin.Fd())
		_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(oldState)), 0, 0, 0)
		if errno != 0 {
			fmt.Printf("Error resetting terminal attributes: %v\n", errno)
		}
	}
}
