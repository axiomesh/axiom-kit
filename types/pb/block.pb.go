// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.4
// source: block.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Header       *BlockHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Transactions [][]byte     `protobuf:"bytes,2,rep,name=transactions,proto3" json:"transactions,omitempty"`
	Extra        *BlockExtra  `protobuf:"bytes,3,opt,name=extra,proto3" json:"extra,omitempty"`
}

func (x *Block) Reset() {
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_block_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_block_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_block_proto_rawDescGZIP(), []int{0}
}

func (x *Block) GetHeader() *BlockHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Block) GetTransactions() [][]byte {
	if x != nil {
		return x.Transactions
	}
	return nil
}

func (x *Block) GetExtra() *BlockExtra {
	if x != nil {
		return x.Extra
	}
	return nil
}

type BlockExtra struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Size int64 `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *BlockExtra) Reset() {
	*x = BlockExtra{}
	if protoimpl.UnsafeEnabled {
		mi := &file_block_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockExtra) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockExtra) ProtoMessage() {}

func (x *BlockExtra) ProtoReflect() protoreflect.Message {
	mi := &file_block_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockExtra.ProtoReflect.Descriptor instead.
func (*BlockExtra) Descriptor() ([]byte, []int) {
	return file_block_proto_rawDescGZIP(), []int{1}
}

func (x *BlockExtra) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

type BlockHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Number         uint64 `protobuf:"varint,1,opt,name=number,proto3" json:"number,omitempty"`
	StateRoot      []byte `protobuf:"bytes,2,opt,name=state_root,json=stateRoot,proto3" json:"state_root,omitempty"`
	TxRoot         []byte `protobuf:"bytes,3,opt,name=tx_root,json=txRoot,proto3" json:"tx_root,omitempty"`
	ReceiptRoot    []byte `protobuf:"bytes,4,opt,name=receipt_root,json=receiptRoot,proto3" json:"receipt_root,omitempty"`
	ParentHash     []byte `protobuf:"bytes,5,opt,name=parent_hash,json=parentHash,proto3" json:"parent_hash,omitempty"`
	Timestamp      int64  `protobuf:"varint,7,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Epoch          uint64 `protobuf:"varint,8,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Bloom          []byte `protobuf:"bytes,9,opt,name=bloom,proto3" json:"bloom,omitempty"`
	ProposerNodeId uint64 `protobuf:"varint,10,opt,name=proposer_node_id,json=proposerNodeId,proto3" json:"proposer_node_id,omitempty"`
	GasPrice       uint64 `protobuf:"varint,11,opt,name=gas_price,json=gasPrice,proto3" json:"gas_price,omitempty"`
	GasUsed        uint64 `protobuf:"varint,12,opt,name=gas_used,json=gasUsed,proto3" json:"gas_used,omitempty"`
	TotalGasFee    []byte `protobuf:"bytes,13,opt,name=total_gas_fee,json=totalGasFee,proto3" json:"total_gas_fee,omitempty"`
	GasFeeReward   []byte `protobuf:"bytes,14,opt,name=gas_fee_reward,json=gasFeeReward,proto3" json:"gas_fee_reward,omitempty"`
}

func (x *BlockHeader) Reset() {
	*x = BlockHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_block_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockHeader) ProtoMessage() {}

func (x *BlockHeader) ProtoReflect() protoreflect.Message {
	mi := &file_block_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockHeader.ProtoReflect.Descriptor instead.
func (*BlockHeader) Descriptor() ([]byte, []int) {
	return file_block_proto_rawDescGZIP(), []int{2}
}

func (x *BlockHeader) GetNumber() uint64 {
	if x != nil {
		return x.Number
	}
	return 0
}

func (x *BlockHeader) GetStateRoot() []byte {
	if x != nil {
		return x.StateRoot
	}
	return nil
}

func (x *BlockHeader) GetTxRoot() []byte {
	if x != nil {
		return x.TxRoot
	}
	return nil
}

func (x *BlockHeader) GetReceiptRoot() []byte {
	if x != nil {
		return x.ReceiptRoot
	}
	return nil
}

func (x *BlockHeader) GetParentHash() []byte {
	if x != nil {
		return x.ParentHash
	}
	return nil
}

func (x *BlockHeader) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *BlockHeader) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *BlockHeader) GetBloom() []byte {
	if x != nil {
		return x.Bloom
	}
	return nil
}

func (x *BlockHeader) GetProposerNodeId() uint64 {
	if x != nil {
		return x.ProposerNodeId
	}
	return 0
}

func (x *BlockHeader) GetGasPrice() uint64 {
	if x != nil {
		return x.GasPrice
	}
	return 0
}

func (x *BlockHeader) GetGasUsed() uint64 {
	if x != nil {
		return x.GasUsed
	}
	return 0
}

func (x *BlockHeader) GetTotalGasFee() []byte {
	if x != nil {
		return x.TotalGasFee
	}
	return nil
}

func (x *BlockHeader) GetGasFeeReward() []byte {
	if x != nil {
		return x.GasFeeReward
	}
	return nil
}

type BlockBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Transactions [][]byte    `protobuf:"bytes,1,rep,name=transactions,proto3" json:"transactions,omitempty"`
	Extra        *BlockExtra `protobuf:"bytes,2,opt,name=extra,proto3" json:"extra,omitempty"`
}

func (x *BlockBody) Reset() {
	*x = BlockBody{}
	if protoimpl.UnsafeEnabled {
		mi := &file_block_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockBody) ProtoMessage() {}

func (x *BlockBody) ProtoReflect() protoreflect.Message {
	mi := &file_block_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockBody.ProtoReflect.Descriptor instead.
func (*BlockBody) Descriptor() ([]byte, []int) {
	return file_block_proto_rawDescGZIP(), []int{3}
}

func (x *BlockBody) GetTransactions() [][]byte {
	if x != nil {
		return x.Transactions
	}
	return nil
}

func (x *BlockBody) GetExtra() *BlockExtra {
	if x != nil {
		return x.Extra
	}
	return nil
}

var File_block_proto protoreflect.FileDescriptor

var file_block_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70,
	0x62, 0x22, 0x7a, 0x0a, 0x05, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x27, 0x0a, 0x06, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x62, 0x2e,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x12, 0x22, 0x0a, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x24, 0x0a, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x45, 0x78, 0x74, 0x72, 0x61, 0x52, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61, 0x22, 0x20, 0x0a,
	0x0a, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x45, 0x78, 0x74, 0x72, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x73,
	0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22,
	0x97, 0x03, 0x0a, 0x0b, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12,
	0x16, 0x0a, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x74, 0x61,
	0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x78, 0x5f, 0x72, 0x6f, 0x6f,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x78, 0x52, 0x6f, 0x6f, 0x74, 0x12,
	0x21, 0x0a, 0x0c, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x6f,
	0x6f, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x68, 0x61, 0x73,
	0x68, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x48,
	0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x6c, 0x6f, 0x6f, 0x6d,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x6f, 0x6d, 0x12, 0x28, 0x0a,
	0x10, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x72, 0x5f, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0e, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65,
	0x72, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x67, 0x61, 0x73, 0x5f, 0x70,
	0x72, 0x69, 0x63, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x67, 0x61, 0x73, 0x50,
	0x72, 0x69, 0x63, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x61, 0x73, 0x5f, 0x75, 0x73, 0x65, 0x64,
	0x18, 0x0c, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x67, 0x61, 0x73, 0x55, 0x73, 0x65, 0x64, 0x12,
	0x22, 0x0a, 0x0d, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x67, 0x61, 0x73, 0x5f, 0x66, 0x65, 0x65,
	0x18, 0x0d, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x47, 0x61, 0x73,
	0x46, 0x65, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x67, 0x61, 0x73, 0x5f, 0x66, 0x65, 0x65, 0x5f, 0x72,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x67, 0x61, 0x73,
	0x46, 0x65, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x22, 0x55, 0x0a, 0x09, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x22, 0x0a, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0c, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x24, 0x0a, 0x05, 0x65, 0x78,
	0x74, 0x72, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x45, 0x78, 0x74, 0x72, 0x61, 0x52, 0x05, 0x65, 0x78, 0x74, 0x72, 0x61,
	0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_block_proto_rawDescOnce sync.Once
	file_block_proto_rawDescData = file_block_proto_rawDesc
)

func file_block_proto_rawDescGZIP() []byte {
	file_block_proto_rawDescOnce.Do(func() {
		file_block_proto_rawDescData = protoimpl.X.CompressGZIP(file_block_proto_rawDescData)
	})
	return file_block_proto_rawDescData
}

var file_block_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_block_proto_goTypes = []interface{}{
	(*Block)(nil),       // 0: pb.Block
	(*BlockExtra)(nil),  // 1: pb.BlockExtra
	(*BlockHeader)(nil), // 2: pb.BlockHeader
	(*BlockBody)(nil),   // 3: pb.BlockBody
}
var file_block_proto_depIdxs = []int32{
	2, // 0: pb.Block.header:type_name -> pb.BlockHeader
	1, // 1: pb.Block.extra:type_name -> pb.BlockExtra
	1, // 2: pb.BlockBody.extra:type_name -> pb.BlockExtra
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_block_proto_init() }
func file_block_proto_init() {
	if File_block_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_block_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Block); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_block_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockExtra); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_block_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockHeader); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_block_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockBody); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_block_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_block_proto_goTypes,
		DependencyIndexes: file_block_proto_depIdxs,
		MessageInfos:      file_block_proto_msgTypes,
	}.Build()
	File_block_proto = out.File
	file_block_proto_rawDesc = nil
	file_block_proto_goTypes = nil
	file_block_proto_depIdxs = nil
}
