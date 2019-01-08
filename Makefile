TARGET=git-repo


git-repo:
	go build

test: git-repo
	golint ./...
	go test ./...
	make -C test

clean:
	rm -f $(TARGET)

.PHONY: test clean
