.PHONY: config

clean:
	rm -f cabby

cert:
	openssl req -x509 -newkey rsa:4096 -nodes -keyout server.key -out server.crt -days 365 -subj "/C=US/ST=Maryland/L=Baltimore/O=Cabby TAXII 2.0/CN=pladdy"
	chmod 600 server.key

config:
	for file in $(shell find config/*example.json -type f | sed 's/.example.json//'); do \
		cp $${file}.example.json $${file}.json; \
	done

cover:
	go test -v -coverprofile=cover.out
	go tool cover -func=cover.out
	@echo
	@echo "'make cover html=true' to see coverage details in a browser"
ifeq ("$(html)","true")
	go tool cover -html=cover.out
endif

deps:
	go get gopkg.in/gorethink/gorethink.v3

fmt:
	go fmt -x

run:
	ls *.go | grep -v _test | xargs go run

test:
	go test -v -cover ./...
