all: tests

tests:
	go build -x -o ./test/fileserver/fileserver ./test/fileserver
	go build -x -o ./test/ping/ping ./test/ping
	go test -v

clean:
	rm -f ./bam ./test/fileserver/fileserver ./test/ping/ping