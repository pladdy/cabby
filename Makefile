.PHONY: config test

GO_FILES=$(shell find .  -name '*go' | grep -v test)

all: config cert build

build:
	go build -o bin/cabby $(GO_FILES)

clean:
	rm -rf bin/
	rm -f cover.out
	rm -f *.log

cert:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout server.key -out server.crt -days 365 -subj "/C=US/ST=Maryland/L=Baltimore/O=Cabby TAXII 2.0/CN=pladdy"
	chmod 600 server.key

config:
	for file in $(shell find config/*example.json -type f | sed 's/.example.json//'); do \
		cp $${file}.example.json $${file}.json; \
	done
	@echo Configs available in config/

cover:
	go test -v -coverprofile=cover.out
	go tool cover -func=cover.out
	@echo
	@echo "'make cover html=true' to see coverage details in a browser"
ifeq ("$(html)","true")
	go tool cover -html=cover.out
endif

fmt:
	go fmt -x

run:
	go run $(GO_FILES)

test:
	go test -v -cover ./...
