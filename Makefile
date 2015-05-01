all: tests

deps:
	go get github.com/BurntSushi/toml
	go get github.com/jweslley/procker
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

server:
	./bam -config examples/bam.conf

clean:
	rm -f ./bam ./examples/fileserver/fileserver ./examples/ping/ping
