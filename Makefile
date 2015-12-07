default: build

build: fix
	go build -v .

buildtests: fix
	cd test/generate && make
	test/generate/generate

test: buildtests quicktest

quicktest: fix
	go test -v

fix: *.go
	goimports -l -w .
	gofmt -l -w .
