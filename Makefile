VERSION=0.1.0

all: tests

deps:
	go get ./...

build: deps
	go build

tests: build
	go test -v

qa: build
	go vet
	golint
	go test -coverprofile=.localdns.cover~
	go tool cover -html=.localdns.cover~

dist:
	packer --os linux  --arch amd64 --output localdns-linux-amd64-$(VERSION).zip
	rm localdns
	packer --os linux  --arch 386   --output localdns-linux-386-$(VERSION).zip
	rm localdns
	packer --os darwin --arch amd64 --output localdns-mac-amd64-$(VERSION).zip
	rm localdns
	packer --os darwin --arch 386   --output localdns-mac-386-$(VERSION).zip
	rm localdns

clean:
	rm -f ./localdns *.zip
