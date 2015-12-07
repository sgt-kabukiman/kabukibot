default: build

build: fix
	go build -v .

test: fix
	cd test/generate && make
	test/generate/generate
	go test -v

quicktest: fix
	go test -v

fix: *.go
	goimports -l -w .
	gofmt -l -w .
