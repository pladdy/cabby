.PHONY: all build clean config cover cover-html fmt reportcard run run-log sqlite
.PHONY: test test-failures test-install test test-run

GO_FILES=$(shell find pkg/ -name '*go' | grep -v test)
BUILD_TAGS=-tags json1
BUILD_PATH=build/cabby

all: config cert dependencies

build:
	go build $(BUILD_TAGS) -o $(BUILD_PATH) $(GO_FILES)

build/debian/etc/cabby/:
	mkdir -p $@

build/debian/etc/systemd/:
	mkdir -p $@

build/debian/lib/systemd/system/cabby.service.d/:
	mkdir -p $@

build/debian/usr/bin/:
	mkdir -p $@

build/debian/var/cabby/:
	mkdir -p $@

build-debian: config build/debian/etc/cabby/
	cp config/cabby.json build/debian/etc/cabby/
	vagrant up
	@echo Magic has happend to make a debian...

clean:
	rm -rf db/
	rm -f server.key server.crt *.log cover.out config/cabby.json

cert:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout server.key -out server.crt -days 365 -subj "/C=US/O=Cabby TAXII 2.0/CN=pladdy"
	chmod 600 server.key

config:
	@for file in $(shell find config/*example.json -type f | sed 's/.example.json//'); do \
		cp $${file}.example.json $${file}.json; \
	done
	@echo Configs available in config/

cover: test-install
ifdef pkg
	go test $(BUILD_TAGS) -v -coverprofile=$(pkg).out ./$(pkg)/...
	go tool cover -func=$(pkg).out
	rm $(pkg).out
else
	@for package in sqlite http; do \
		go test $(BUILD_TAGS) -v -coverprofile=$${package}.out ./$${package}/...; \
		go tool cover -func=$${package}.out; \
		rm $${package}.out; \
	done
endif

cover-html: test-install
ifdef pkg
	go test $(BUILD_TAGS) -v -coverprofile=$(pkg).out ./$(pkg)/...
	go tool cover -func=$(pkg).out
	go tool cover -html=$(pkg).out
	rm $(pkg).out
else
	@for package in sqlite http; do \
		go test $(BUILD_TAGS) -v -coverprofile=$${package}.out ./$${package}/...; \
		go tool cover -func=$${package}.out; \
		go tool cover -html=$${package}.out; \
		rm $${package}.out; \
	done
endif

dependencies:
	go get -t -v  ./...
	go get github.com/fzipp/gocyclo
	go get github.com/golang/lint

dev-db:
	build/debian/usr/bin/cabby-cli -u test@cabby.com -p test -a

fmt:
	go fmt -x

reportcard: fmt
	gocyclo -over 10 .
	golint
	go vet

run:
	go run $(BUILD_TAGS) $(GO_FILES)

run-log:
	go run $(BUILD_TAGS) $(GO_FILES) 2>&1 | tee cabby.log

test: test-install
ifdef pkg
	go test $(BUILD_TAGS) -v ./$(pkg)/...
else
	go test $(BUILD_TAGS) -v -cover ./...
endif

test-failures: test-install
	go test $(BUILD_TAGS) -v ./... 2>&1 | grep -A 1 FAIL

test-install:
	go test $(BUILD_TAGS) -i ./

test-run: test-install
ifdef test
	go test $(BUILD_TAGS) -v ./... -run $(test)
else
	@echo Syntax is 'make $@ test=<test name>'
endif
