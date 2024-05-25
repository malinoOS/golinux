parentFolder := $(shell pwd)
DRIVE = nbd0
SHELL := /bin/bash

all: clean prepare buildInit buildFallsh install qemu

createVM:
	sudo modprobe nbd max_part=8
	qemu-img create -f qcow2 linux.qcow2 16G
	sudo qemu-nbd -c /dev/$(DRIVE) linux.qcow2
	echo -e "o\nn\np\n\n\n\nw" | sudo fdisk /dev/$(DRIVE)
	sudo mkfs -t ext4 /dev/$(DRIVE)p1
	mkdir disk
	sudo mount -t ext4 /dev/$(DRIVE)p1 disk
	sudo mkdir -pv disk/{bin,sbin,etc,lib,lib64,var,dev,proc,sys,run,tmp,boot}
	sudo mknod -m 600 disk/dev/console c 5 1
	sudo mknod -m 600 disk/dev/tty c 5 1
	sudo mknod -m 666 disk/dev/null c 1 3
	sudo cp $$(ls -t /boot/vmlinuz* | head -n1) disk/boot/
	sudo cp $$(ls -t /boot/initrd* | head -n1) disk/boot/
	sudo grub-install /dev/$(DRIVE) --skip-fs-probe --boot-directory=disk/boot --target=i386-pc
	sudo printf "set default=0\nset timeout=1\n\nmenuentry \"golinux\" {\n    linux $$(ls -t /boot/vmlinuz* | head -n1) root=/dev/sda1 ro\n    initrd $$(ls -t /boot/initrd* | head -n1)\n}" | sudo tee disk/boot/grub/grub.cfg
	sudo printf "[init]\nprintSplashMessage = true\nremountRootPartitionAsWritable = true\nmalinoMode = true\nexec = /bin/fallsh" | sudo tee disk/etc/init.ini
	sudo umount disk
	rm -rf disk
	sudo qemu-nbd -d /dev/$(DRIVE)
	#sudo rmmod nbd
	
undoVM:
	-sudo umount disk
	-rm -rf disk
	-sudo qemu-nbd -d /dev/$(DRIVE)
	-rm linux.qcow2
	
clean:
	rm -rf bin/

prepare:
	mkdir bin

buildInit:
	cd $(parentFolder)/init; \
	go mod tidy; \
	go build -o $(parentFolder)/bin/init -ldflags "-X main.Version=$(shell date +%y%m%d)"
	
buildFallsh:
	cd $(parentFolder)/fallsh; \
	go mod tidy; \
	go build -o $(parentFolder)/bin/fallsh -ldflags "-X main.Version=$(shell date +%y%m%d)"
	
#buildGkilo:
#	cd $(parentFolder)/gkilo/src; \
#	go mod tidy; \
#	go build -o $(parentFolder)/bin/gkilo

install:
	sudo modprobe nbd max_part=8
	sudo qemu-nbd -c /dev/$(DRIVE) linux.qcow2
	mkdir -p disk
	sudo mount -t ext4 /dev/$(DRIVE)p1 disk
	sudo cp $(parentFolder)/bin/init disk/sbin/init
	sudo cp $(parentFolder)/bin/fallsh disk/bin/fallsh
	#sudo cp $(parentFolder)/bin/gkilo disk/bin/gkilo
	sudo umount disk
	rm -rf disk
	sudo qemu-nbd -d /dev/$(DRIVE)
	
qemu:
	qemu-system-x86_64 -drive file=linux.qcow2,format=qcow2 -m 4G -enable-kvm -smp 4