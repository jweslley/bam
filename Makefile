PROGRAM=bam
VERSION=0.1.0
LDFLAGS="-X main.programVersion=$(VERSION)"

all: test

deps:
	go get ./...

tools:
	go get github.com/mjibson/esc

server:
	./bam -config examples/bam.conf

build: deps
	go build -o ./examples/fileserver/fileserver ./examples/fileserver
	go build -o ./examples/ping/ping ./examples/ping
	go build

generate:
	go generate

test: build
	go test -v ./...

qa:
	go vet
	golint
	go test -coverprofile=.cover~
	go tool cover -html=.cover~

dist:
	@for os in linux darwin; do \
		for arch in 386 amd64; do \
			target=$(PROGRAM)-$$os-$$arch-$(VERSION); \
			echo Building $$target; \
			GOOS=$$os GOARCH=$$arch go build -ldflags $(LDFLAGS) -o $$target/$(PROGRAM) ; \
			cp ./README.md ./LICENSE $$target; \
			tar -zcf $$target.tar.gz $$target; \
			rm -rf $$target;                   \
		done                                 \
	done

clean:
	rm -rf *.tar.gz
