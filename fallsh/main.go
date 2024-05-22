package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var running bool = true
var console int = 0

func main() {
	fmt.Printf("fallback shell\n")
	// Set up non-canonical mode for reading from /dev/console
	//setNonCanonicalMode()
	//syscall.Open("/dev/console", syscall.O_RDWR|syscall.O_NDELAY, syscall.SYS_IOPERM)

	for running {
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Critcal error while getting working directory:\n%v", err)
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
			fmt.Printf("Critcal error while reading characters:\n%v", err)
			for true {
			}
		}
		if n > 0 {
			char := buf[0]
			if char == '\n' {
				fmt.Println()
				return cmdString.String()
			} else if char == 127 { // ASCII code for backspace
				if cmdString.Len() > 0 {
					// Convert the builder to a string, remove the last character, and create a new builder
					cmd := cmdString.String()
					if len(cmd) > 1 {
						cmdString.Reset()
						cmdString.WriteString(cmd[:len(cmd)-1])
						// Move the cursor back and clear the character
						fmt.Print("\b \b")
					} else {
						// If only one character is present, simply reset the builder
						cmdString.Reset()
						fmt.Print("\b \b") // Clear the character
					}
				}
			} else {
				fmt.Print(string(char))
				cmdString.WriteByte(char)
			}
		} else {
			time.Sleep(time.Millisecond * 10)
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
		fmt.Printf("golinux/gosh commands:\n\nhelp - shows this menu\nexit - exits (also kernel panics)\nreboot - reboots the system\nshutdown - shuts down\ntest - tests functionality\ntasklist - lists processes\ngr - goroutine test\ncd [folder] - change directory\nls - list the current directory\ncat [file] - dump file\n")
	case "reboot":
		fmt.Printf("syncing disks...\n")
		syscall.Sync()
		fmt.Printf("rebooting...\n")
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	case "shutdown":
		fmt.Printf("syncing disks...\n")
		syscall.Sync()
		fmt.Printf("shutting down...\n")
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	case "test":
		fmt.Printf("Starting tests...\n")
		//testing()
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
		currentDir, _ := os.Getwd() // not error checking!11!!!11
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
			currentDir, _ := os.Getwd() // not error checking!11!!!11
			file = fmt.Sprintf("%v/%v", currentDir, arrCommandStr[1])
		}

		dat, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("cat: could not read file: %v\n", err)
			return nil
		}
		fmt.Println(string(dat))
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

func cd(arrCommandStr []string) {
	dir := ""
	if arrCommandStr[1] == ".." {
		// Move to parent directory
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
