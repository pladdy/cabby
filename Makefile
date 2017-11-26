clean:
	rm -f cabby

config:
	cp config.example.json config.json

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

test:
	go test -v -cover ./...
