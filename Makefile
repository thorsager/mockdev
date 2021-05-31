GO_CMD = go
GO_BUILD = $(GO_CMD) build
GO_CLEAN = $(GO_CMD) clean
GO_TEST = $(GO_CMD) test
BIN_PATH = bin
BIN_SNMP_SNAPSHOT = $(BIN_PATH)/snmp-snapshot
BIN_HTTP_DUMP = $(BIN_PATH)/http-dump
BIN_FAKEITD = $(BIN_PATH)/mockdevd
VERSION ?= $(shell git describe --tags --always --dirty 2> /dev/null || echo v0)
LDFLAGS = -w -extldflags -static

.PHONY: all
all: test snmp-snapshot mockdevd http-dump

.PHONY: test
test:
	$(GO_TEST) ./...

.PHONY: snmp-snapshot
snmp-snapshot:
	CGO_ENABLED=0 $(GO_BUILD) -ldflags "-X main.Version=$(VERSION) $(LDFLAGS)" \
 		-o $(BIN_SNMP_SNAPSHOT) \
 		cmd/snmpsnapshot/snmp_snapshot.go

.PHONY: http-dump
http-dump:
	CGO_ENABLED=0 $(GO_BUILD) -ldflags "-X main.Version=$(VERSION) $(LDFLAGS)" \
 		-o $(BIN_HTTP_DUMP) \
 		cmd/httpdump/http_dump.go

.PHONY: mockdevd
mockdevd:
	CGO_ENABLED=0 $(GO_BUILD) -ldflags "-X main.Version=$(VERSION) $(LDFLAGS)" \
		-o $(BIN_FAKEITD) \
		cmd/mockdevd/mockdevd.go

.PHONY: clean
clean:
	$(GO_CLEAN)
	rm -f $(BIN_FAKEITD)
	rm -f $(BIN_SNMP_SNAPSHOT)
	rm -f $(BIN_HTTP_DUMP)
