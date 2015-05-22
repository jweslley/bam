VERSION=0.1.0

all: tests

deps:
	go get ./...
	go get github.com/mjibson/esc

build: deps
	go build -o ./examples/fileserver/fileserver ./examples/fileserver
	go build -o ./examples/ping/ping ./examples/ping
	go build

generate:
	go generate

tests: build
	go test -v

qa: build
	go vet
	golint
	go test -coverprofile=.bam.cover~
	go tool cover -html=.bam.cover~

dist:
	packer --os linux  --arch amd64 --output bam-linux-amd64-$(VERSION).zip
	rm bam
	packer --os linux  --arch 386   --output bam-linux-386-$(VERSION).zip
	rm bam
	packer --os darwin --arch amd64 --output bam-mac-amd64-$(VERSION).zip
	rm bam
	packer --os darwin --arch 386   --output bam-mac-386-$(VERSION).zip
	rm bam

server:
	./bam -config examples/bam.conf

clean:
	rm -f ./bam ./examples/fileserver/fileserver ./examples/ping/ping *.zip
