.PHONY: all build build/debian/usr/bin/cabby build/debian/usr/bin/cabby-cli
.PHONY: clean clean-cli cmd/cabby-cli/cabby-cli cover cover-html db/cabby.db reportcard run run-log
.PHONY: test test test-run vagrant

BUILD_TAGS=-tags json1
BUILD_PATH=build/cabby
CLI_FILES=$(shell find cmd/cabby-cli/*.go -name '*go' | grep -v test)
PACKAGES=./ sqlite http cmd/cabby-cli

all: config/cabby.json cert dependencies

build: dependencies build/debian/usr/bin/cabby build/debian/usr/bin/cabby-cli

build/debian/etc/cabby/:
	mkdir -p $@

build/debian/etc/systemd/:
	mkdir -p $@

build/debian/lib/systemd/system/cabby.service.d/:
	mkdir -p $@

build/debian/usr/bin/:
	mkdir -p $@

build/debian/usr/bin/cabby-cli: build/debian/usr/bin/
	go build $(BUILD_TAGS) -o $@ $(CLI_FILES)

build/debian/usr/bin/cabby: build/debian/usr/bin/ build/debian/etc/cabby/cabby.json build/debian/var/cabby/schema.sql
	go build $(BUILD_TAGS) -o $@ cmd/cabby/main.go

build/debian/var/cabby/:
	mkdir -p $@

build/debian/var/cabby/schema.sql: sqlite/schema.sql build/debian/var/cabby/
	cp $< $@

build-debian:
	vagrant up
	vagrant provision --provision-with build-cabby
	@echo Magic has happend to make a debian...
	vagrant destroy -f

cabby.deb: build
	fpm -f \
		-s dir \
		-t deb \
		-n cabby \
		-p $@ \
		-m "Matt Pladna" \
		--description "A TAXII 2.0 server" \
		--after-install build/debian/postinst \
		--deb-user cabby \
		--deb-group cabby \
		--deb-pre-depends make \
		--deb-pre-depends jq \
		--deb-pre-depends sqlite \
		-C build/debian .

clean:
	rm -rf db/cabby.db
	rm -f server.key server.crt *.log cover.out config/cabby.json
	rm -f build/debian/usr/bin/cabby build/debian/usr/bin/cabby-cli

clean-cli:
	rm -f build/debian/usr/bin/cabby-cli

cert:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout server.key -out server.crt -days 365 -subj "/C=US/O=Cabby TAXII 2.0/CN=pladdy"
	chmod 600 server.key

cmd/cabby-cli/cabby-cli:
	go build $(BUILD_TAGS) -o $@ $(CLI_FILES)

config/cabby.json: config/cabby.example.json
	cp $< $@

cover:
ifdef pkg
	go test $(BUILD_TAGS) -i ./$(pkg)
	go test $(BUILD_TAGS) -v -failfast -coverprofile=$(pkg).out ./$(pkg)
	go tool cover -func=$(pkg).out
else
	@for package in $(PACKAGES); do \
	  go test $(BUILD_TAGS) -i ./$${package}; \
		go test $(BUILD_TAGS) -v -failfast -coverprofile=$${package}.out ./$${package}; \
		go tool cover -func=$${package}.out; \
	done
endif

cover-html:
ifdef pkg
	$(MAKE) cover pkg=$(pkg)
	go tool cover -html=$(pkg).out
else
	$(MAKE) cover
	@for package in $(PACKAGES); do \
		go tool cover -html=$${package}.out; \
	done
endif

cover-cabby.txt:
	go test -v $(BUILD_TAGS) -failfast -coverprofile=$@ -covermode=atomic ./ ./

cover-http.txt:
	go test -v $(BUILD_TAGS) -failfast -coverprofile=$@ -covermode=atomic ./http/...

cover-sqlite.txt:
	go test -v $(BUILD_TAGS) -failfast -coverprofile=$@ -covermode=atomic ./sqlite/...

coverage.txt: cover-cabby.txt cover-http.txt cover-sqlite.txt
	@cat $^ > $@
	@rm -f $^

db/cabby.db: cmd/cabby-cli/cabby-cli
	cmd/local-db

dependencies:
	go get -t -v  ./...
	go get github.com/fzipp/gocyclo
	go get golang.org/x/lint/golint

dev-db: db/cabby.db

fmt:
	go fmt -x

reportcard: fmt
	gocyclo -over 10 .
	golint
	go vet

run:
	go run $(BUILD_TAGS) cmd/cabby/main.go -config config/cabby.json

run-cli:
	go run $(CLI_FILES)

run-log:
	go run $(BUILD_TAGS) cmd/cabby/main.go 2>&1 | tee cabby.log

test:
ifdef pkg
	go test $(BUILD_TAGS) -i ./$(pkg)
	go test $(BUILD_TAGS) -v -failfast -cover ./$(pkg)
else
	go test $(BUILD_TAGS) -i ./...
	go test $(BUILD_TAGS) -v -failfast -cover ./...
endif

test-run:
ifdef test
	go test $(BUILD_TAGS) -i ./...
	go test $(BUILD_TAGS) -v -failfast ./... -run $(test)
else
	@echo Syntax is 'make $@ test=<test name>'
endif

vagrant:
	vagrant up
