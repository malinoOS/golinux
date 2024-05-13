parentFolder := $(shell pwd)

all: clean prepare buildInit install qemu

clean:
	rm -rf bin/

prepare:
	mkdir bin

buildInit:
	cd $(parentFolder)/init; \
	go mod tidy; \
	go build -o $(parentFolder)/bin/init
	
install:
	sudo mount /dev/sdc1 /mnt
	sudo cp $(parentFolder)/bin/init /mnt/sbin/init
	sudo umount /mnt
	
qemu:
	sudo qemu-system-x86_64 -hda /dev/sdc -m 4G -enable-kvm