.PHONY: all build clean config cover cover-html fmt reportcard run run-log sqlite
.PHONY: test test-failures test-install test test-run

BUILD_TAGS=-tags json1
BUILD_PATH=build/cabby
PACKAGES=./ sqlite/... http/...

all: config cert dependencies

build: build/debian/usr/bin/
	go build $(BUILD_TAGS) -o build/debian/usr/bin/cabby cmd/cabby/main.go

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
	go test $(BUILD_TAGS) -i ./$(pkg)
	go test $(BUILD_TAGS) -v -coverprofile=$(pkg).out ./$(pkg)
	go tool cover -func=$(pkg).out
	rm $(pkg).out
else
	@for package in $(PACKAGES); do \
	  go test $(BUILD_TAGS) -i ./$${package}; \
		go test $(BUILD_TAGS) -v -coverprofile=$${package}.out ./$${package}; \
		go tool cover -func=$${package}.out; \
		rm $${package}.out; \
	done
endif

cover-html: test-install
ifdef pkg
	go test $(BUILD_TAGS) -i ./$(pkg)
	go test $(BUILD_TAGS) -v -coverprofile=$(pkg).out ./$(pkg)
	go tool cover -func=$(pkg).out
	go tool cover -html=$(pkg).out
	rm $(pkg).out
else
	@for package in $(PACKAGES); do \
	  go test $(BUILD_TAGS) -i ./$${package}; \
		go test $(BUILD_TAGS) -v -coverprofile=$${package}.out ./$${package}; \
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
	cmd/cabby-cli -u test@cabby.com -p test -a

fmt:
	go fmt -x

reportcard: fmt
	gocyclo -over 10 .
	golint
	go vet

run:
	go run $(BUILD_TAGS) cmd/cabby/main.go

run-log:
	go run $(BUILD_TAGS) cmd/cabby/main.go 2>&1 | tee cabby.log

test:
ifdef pkg
	go test $(BUILD_TAGS) -i ./$(pkg)
	go test $(BUILD_TAGS) -v -cover ./$(pkg)
else
	go test $(BUILD_TAGS) -v -cover ./...
endif

test-failures:
	go test $(BUILD_TAGS) -v ./... 2>&1 | grep -A 1 FAIL

test-run:
ifdef test
	go test $(BUILD_TAGS) -i ./...
	go test $(BUILD_TAGS) -v ./... -run $(test)
else
	@echo Syntax is 'make $@ test=<test name>'
endif
