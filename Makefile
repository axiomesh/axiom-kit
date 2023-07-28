GO_BIN = go
ifneq (${GO},)
	GO_BIN = ${GO}
endif

GO_SRC_PATH := $(GOPATH)/src
CURRENT_PATH = $(shell pwd)/types/pb

help: Makefile
	@printf "${BLUE}Choose a command run:${NC}\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/    /'

## make test: Run go unittest
test:
	${GO_BIN} test ./...

## make linter: Run golanci-lint
linter:
	golangci-lint run -E goimports -E bodyclose --skip-dirs-use-default

install-dependences:
	${GO_BIN} install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	${GO_BIN} install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@v0.4.0

clean-pb:
	rm -rf $(CURRENT_PATH)/*.pb.go

compile-pb: clean-pb
	protoc --proto_path=$(CURRENT_PATH) -I. -I$(GO_SRC_PATH) --go_out=$(CURRENT_PATH) --plugin protoc-gen-go="$(GOPATH)/bin/protoc-gen-go" --go-vtproto_out=$(CURRENT_PATH) --plugin protoc-gen-go-vtproto="$(GOPATH)/bin/protoc-gen-go-vtproto" --go-vtproto_opt=features=marshal+marshal_strict+unmarshal+size $(CURRENT_PATH)/*.proto

.PHONY: install-dependences clean-pb compile-pb