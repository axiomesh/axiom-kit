GO_BIN = go
ifneq (${GO},)
	GO_BIN = ${GO}
endif

GO_SRC_PATH = $(GOPATH)/src
CURRENT_PATH = $(shell pwd)/types/pb
PB_PKG_PATH = ../pb

help: Makefile
	@printf "${BLUE}Choose a command run:${NC}\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/    /'

## make test: Run go unittest
test:
	${GO_BIN} test ./...

## make linter: Run golanci-lint
linter:
	golangci-lint run -E goimports -E bodyclose --skip-dirs-use-default

prepare:
	${GO_BIN} install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	${GO_BIN} install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@v0.5

clean-pb:
	rm -rf $(CURRENT_PATH)/*.pb.go

compile-pb: clean-pb
	protoc --proto_path=$(CURRENT_PATH) -I. -I$(GO_SRC_PATH) \
		--go_out=$(CURRENT_PATH) \
		--plugin protoc-gen-go="$(GOPATH)/bin/protoc-gen-go" \
		--go-vtproto_out=$(CURRENT_PATH) \
		--plugin protoc-gen-go-vtproto="$(GOPATH)/bin/protoc-gen-go-vtproto" \
		--go-vtproto_opt=features=marshal+marshal_strict+unmarshal+size+equal+pool \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).BytesSlice \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Block \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).BlockHeader \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).BlockExtra \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).BlockBody \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).ChainMeta \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).TransactionMeta \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Message \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Receipt \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Receipts \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Event \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).EvmLog \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Message \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).SyncStateRequest \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).SyncStateResponse \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).SyncBlockRequest \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).SyncChainDataRequest \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).SyncChainDataResponse \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).FetchEpochStateRequest \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).Node \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).InternalNode \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).LeafNode \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).InnerAccount \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).TrieJournal \
		--go-vtproto_opt=pool=$(PB_PKG_PATH).TrieJournalBatch \
		$(CURRENT_PATH)/*.proto

.PHONY: prepare clean-pb compile-pb