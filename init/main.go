package main

import (
	"fmt"
	"os"

	//"os/exec"
	"os/signal"
	"syscall"

	"time"

	"github.com/go-ini/ini"
	"golang.org/x/sys/unix"
)

var cfg ini.File

func main() {
	fmt.Printf("\033[97mReading configuration... ")
	cfg, err := ini.Load("/etc/init.ini")
	if err != nil {
		panicScreen(err)
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
			panicScreen(err)
		} else {
			fmt.Printf("\033[92m[OK]\033[39m\n")
		}
	}

	fmt.Printf("Creating stdio... ")
	fstdin, err0 := os.Create("/dev/stdin")
	fstdout, err1 := os.Create("/dev/stdout")
	fstderr, err2 := os.Create("/dev/stderr")
	if err0 != nil && err1 != nil && err2 != nil {
		panicScreen(err0)
		for true {
		}
	} else {
		fmt.Printf("\033[92m[OK]\033[39m\n")
	}

	fmt.Printf("Spawning the Fallback Shell (fallsh)\n")

	procAttr := &syscall.ProcAttr{
		Dir:   "/",
		Env:   []string{"OSENV=malino"},
		Files: []uintptr{fstdin.Fd(), fstdout.Fd(), fstderr.Fd()},
		Sys:   nil,
	}

	//exec.Command("/bin/gosh", "")
	//syscall.Exec("/bin/gosh", []string{"/bin/gosh"}, os.Environ())

	var wstatus syscall.WaitStatus

	pid, err := syscall.ForkExec("/sbin/malino", nil, procAttr)
	if err != nil {
		fmt.Printf("err: could not execute malino OS")
	} else {
		_, err = syscall.Wait4(pid, &wstatus, 0, nil)
		if err != nil {
			fmt.Printf("err: could not execute malino OS")
		}
	}

	if wstatus.Exited() {
		// Process exited
		// Create a new error
		fmt.Printf("err: malino OS exited with code %d", wstatus.ExitStatus())
	}

	pid, err = syscall.ForkExec("/bin/fallsh", nil, procAttr)
	if err != nil {
		panicScreen(err)
	} else {
		_, err = syscall.Wait4(pid, &wstatus, 0, nil)
		if err != nil {
			panicScreen(err)
		}
	}

	if wstatus.Exited() {
		// Process exited
		// Create a new error
		panicScreen(fmt.Errorf("fallback shell crashed. Exit code: %d", wstatus.ExitStatus()))
	}

	for true {
	}
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt,
		syscall.SIGCHLD,
		syscall.SIGTERM,
		syscall.SIGKILL)

	go func() {
		for sig := range c {
			fmt.Printf("captured %v", sig)
		}
	}()
}

func panicScreen(err error) {
	fmt.Print("\033[104m\033[H\033[J")
	fmt.Printf("A problem has been detected and golinux has been frozen because program execution literally cannot continue.\n\n%v\n\n", err.Error())
	fmt.Println("If this is the first time you've seen this error screen, restart your computer. If this screen appears again, follow these steps:\n")
	fmt.Println("Check to make sure that the configuration file \"/etc/init.ini\" exists and is correct.\n")
	fmt.Println("If problems continue, on another device, go to https://github.com/malinoOS/golinux. Click on the issues tab, and click \"New issue\". From there,")
	fmt.Println("Write an accurate description of your problem and submit the issue. You should get a response within the next couple of hours or days.\n")

	fmt.Print("Shutting down in 10")
	time.Sleep(time.Second)
	fmt.Print("\b \b\b \b9")
	time.Sleep(time.Second)
	i := 9
	for i != 0 {
		i--
		fmt.Printf("\b%v", i)
		time.Sleep(time.Second)
	}
	fmt.Printf("syncing disks...\n")
	unix.Sync()
	fmt.Printf("shutting down...\n")
	unix.Reboot(unix.LINUX_REBOOT_CMD_POWER_OFF)
}
