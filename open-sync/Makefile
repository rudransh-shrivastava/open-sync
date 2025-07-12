DAEMON_BINARY_PATH= bin/daemon
PROTO_FILE = internal/protobuf/protocol.proto
PROTO_GO = internal/protobuf/protocol.pb.go

# Protocol buffer generation
.PHONY: protoc
protoc: $(PROTO_GO)

$(PROTO_GO): $(PROTO_FILE)
	@echo "Generating protocol buffers..."
	@protoc --go_out=. --go_opt=paths=source_relative $<

# Daemon targets
.PHONY: build
build: protoc
		@echo "Building daemon..."
		@mkdir -p bin
		@go build -o $(DAEMON_BINARY_PATH) ./cmd/daemon

.PHONY: run
run: build
		@echo "Running daemon..."
		@./$(DAEMON_BINARY_PATH)
