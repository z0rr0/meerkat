PROGRAM=MEERKAT
BIN=bin
ROOTPKG=github.com/z0rr0/meerkat
VERSION=`bash version.sh $(ROOTPKG)`
SOURCEDIR=$(GOPATH)/src/$(ROOTPKG)
TMPDIR=/tmp


all: test

install:
	go install -ldflags "$(VERSION)" $(ROOTPKG)/server
	go install -ldflags "$(VERSION)" $(ROOTPKG)/client
	go install -ldflags "$(VERSION)" $(ROOTPKG)/packet

configs:
	cp -f $(SOURCEDIR)/server/meerkat.json $(TMPDIR)/meerkat.json

run_server: configs install
	$(GOPATH)/$(BIN)/server

lint: install
	golint $(ROOTPKG)/packet
	golint $(ROOTPKG)/client
	golint $(ROOTPKG)/server

test: lint
	# go tool cover -html=coverage.out
	# go tool trace ratest.test trace.out
	go test -race -v -cover -coverprofile=server_coverage.out -trace server_trace.out $(ROOTPKG)/server
	go test -race -v -cover -coverprofile=client_coverage.out -trace client_trace.out $(ROOTPKG)/client

bench: lint
	go test -bench=. -benchmem -v $(ROOTPKG)/packet
	go test -bench=. -benchmem -v $(ROOTPKG)/client
	go test -bench=. -benchmem -v $(ROOTPKG)/client

clean:
	rm -rf $(GOPATH)/$(BIN)/*
	find $(SOURCEDIR) -type f -name "*[coverage,trace].out" -delete
