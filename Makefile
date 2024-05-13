parentFolder := $(shell pwd)

all: clean prepare buildInit buildGosh install qemu

clean:
	rm -rf bin/

prepare:
	mkdir bin

buildInit:
	cd $(parentFolder)/init; \
	go mod tidy; \
	go build -o $(parentFolder)/bin/init
	
buildGosh:
	cd $(parentFolder)/gosh; \
	go mod tidy; \
	go build -o $(parentFolder)/bin/gosh

install:
	sudo mount /dev/sdc1 /mnt
	sudo cp $(parentFolder)/bin/init /mnt/sbin/init
	sudo cp $(parentFolder)/bin/gosh /mnt/bin/gosh
	sudo umount /mnt
	
qemu:
	sudo qemu-system-x86_64 -hda /dev/sdc -m 4G -enable-kvm -smp 4