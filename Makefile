all: tests

tests:
	go get github.com/BurntSushi/toml
	go get github.com/jweslley/procker
	go build -x -o ./test/fileserver/fileserver ./test/fileserver
	go build -x -o ./test/ping/ping ./test/ping
	go test -v

coverage: tests
	go test -coverprofile=bam.cover
	go tool cover -html=bam.cover

clean:
	rm -f ./bam ./test/fileserver/fileserver ./test/ping/ping
