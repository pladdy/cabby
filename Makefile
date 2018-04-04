.PHONY: config test

GO_FILES=$(shell find . -name '*go' | grep -v test)
BUILD_TAGS=-tags json1

all: config cert dependencies

build:
	go build $(BUILD_TAGS) -o bin/cabby $(GO_FILES)

clean:
	rm -rf bin/ db/
	rm -f server.key server.crt *.log cover.out config/cabby.json

cert:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout server.key -out server.crt -days 365 -subj "/C=US/ST=Maryland/L=Baltimore/O=Cabby TAXII 2.0/CN=pladdy"
	chmod 600 server.key

config:
	@for file in $(shell find config/*example.json -type f | sed 's/.example.json//'); do \
		 cp $${file}.example.json $${file}.json; \
	 done
	@echo Configs available in config/

cover: test_install
	go test $(BUILD_TAGS) -v -coverprofile=cover.out
	go tool cover -func=cover.out
	@echo
	@echo "'make cover html=true' to see coverage details in a browser"
ifeq ("$(html)","true")
	go tool cover -html=cover.out
endif
	@rm cover.out

dependencies:
	go get -t -v  ./...
	go get github.com/fzipp/gocyclo
	go get github.com/golang/lint

fmt:
	go fmt -x

reportcard: fmt
	gocyclo -over 10 .
	golint
	go vet

run:
	go run $(BUILD_TAGS) $(GO_FILES)

sqlite:
	rm -rf db/
	mkdir db
	sqlite3 db/cabby.db '.read backend/sqlite/schema.sql'

test: test_install
	go test $(BUILD_TAGS) -v -cover ./...

test_failures: test_install
	go test $(BUILD_TAGS) -v ./... 2>&1 | grep -A 1 FAIL

test_install:
	go test -tags json1 -i

test_run:
ifdef test
	go test $(BUILD_TAGS) -v -run $(test)
else
	@echo Syntax is 'make test_run test=<test name>'
endif
