package main

import (
	"fmt"
	"os"

	//"os/exec"
	"syscall"

	"github.com/go-ini/ini"
)

var cfg ini.File

func main() {
	fmt.Printf("\033[39mReading configuration... ")
	cfg, err := ini.Load("/etc/init.ini")
	if err != nil {
		fmt.Printf("\033[91m[FAIL]\033[39m\nCritical error while reading /etc/init.ini:\n%v\n\nThe system will now freeze.", err.Error())
		for true {
		}
	} else {
		fmt.Printf("\033[92m[OK]\033[39m\n")
	}

	if cfg.Section("init").Key("printSplashMessage").String() == "true" {
		fmt.Printf("\033[96mgolinux!\033[39m\n")
	}

	if cfg.Section("init").Key("remountRootPartitionAsWritable").String() == "true" {
		fmt.Printf("Remounting / as writable... ")
		err := syscall.Mount("", "/", "", syscall.MS_REMOUNT, "")
		if err != nil {
			fmt.Printf("\033[91m[FAIL]\033[39m\nCritical error while remounting /:\n%v\n\nThe system will continue as read-only.", err.Error())
		} else {
			fmt.Printf("\033[92m[OK]\033[39m\n")
		}
	}

	fmt.Printf("Spawning the Go Shell (gosh)...\n")

	//exec.Command("/bin/gosh", "")
	syscall.Exec("/bin/gosh", []string{"/bin/gosh"}, os.Environ())

	for true {
	}

	/*reader := bufio.NewReader(os.Stdin)
	for running {
		fmt.Print("# ")
		cmdString, err := readCommand(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		err = runCommand(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}*/
}

/*func readCommand(reader *bufio.Reader) (string, error) {
	var cmdString strings.Builder
	buffer := make([]byte, 1024) // Larger buffer size
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			return "", err
		}
		for i := 0; i < n; i++ {
			char := buffer[i]
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
	switch arrCommandStr[0] {
	case "exit":
		running = false
		// add another case here for custom commands.
	case "help":
		fmt.Printf("golinux commands:\n\nhelp - shows this menu\nexit - exits (also kernel panics)\nhalt - halts linux\ntest - tests functionality\n")
	case "halt":
		fmt.Printf("syncing disks...\n")
		syscall.Syscall(syscall.SYS_SYNC, 0, 0, 0)
		fmt.Printf("rebooting...\n")
		syscall.Syscall(syscall.SYS_REBOOT, 0xfee1dead, 672274793, 0x1234567)
	case "test":
		fmt.Printf("Starting tests...\n")
		testing()
	default:
		fmt.Printf("invalid command\n")
	}
	return nil
}

func testing() {
	fmt.Printf("File storage... ")
	d := []byte("golinux!\ngo\nlinux!\n")
	err := os.WriteFile("/tmp/golinux.txt", d, 0644)
	if err != nil {
		fmt.Printf("\033[91m[FAILED WRITE]\033[39m\n")
		fmt.Printf("%v", err.Error())
		return
	} else {
		syscall.Syscall(syscall.SYS_SYNC, 0, 0, 0)
		dat, err := os.ReadFile("/tmp/golinux.txt")
		if err != nil {
			fmt.Printf("\033[91m[FAILED READ]\033[39m\n")
			return
		} else {
			if string(dat) == "golinux!\ngo\nlinux!\n" {
				fmt.Printf("\033[92m[PASS]\033[39m\n")
			} else {
				fmt.Printf("\033[91m[FAILED COMPARE]\033[39m\n")
				return
			}
		}
	}

	fmt.Printf("Config read... ")
	dat, err := os.ReadFile("/etc/init.ini")
	if err != nil {
		fmt.Printf("\033[91m[FAILED READ]\033[39m\n")
		fmt.Printf("%v", err.Error())
		return
	} else {
		fmt.Printf("\033[92m[PASS]\033[39m\n")
		fmt.Printf("%v\n", string(dat))
	}
}*/
