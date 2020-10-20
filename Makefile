GO_CMD = go
GO_BUILD = $(GO_CMD) build
GO_CLEAN = $(GO_CMD) clean
GO_TEST = $(GO_CMD) test
BIN_PATH = bin
BIN_SNMP_SNAPSHOT = $(BIN_PATH)/snmp-snapshot
BIN_FAKEITD = $(BIN_PATH)/mockdevd
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)
LDFLAGS = -w -extldflags -static

.PHONY: all
all: snmp-snapshot mockdevd

.PHONY: snmp-snapshot
snmp-snapshot:
	CGO_ENABLED=0 $(GO_BUILD) -ldflags "-X main.Version=$(VERSION) $(LDFLAGS)" \
 		-o $(BIN_SNMP_SNAPSHOT) \
 		cmd/snmpsnapshot/snmp_snapshot.go

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
