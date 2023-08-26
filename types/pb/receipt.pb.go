// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.13.0
// source: receipt.proto

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

type Receipt_Status int32

const (
	Receipt_SUCCESS Receipt_Status = 0
	Receipt_FAILED  Receipt_Status = 1
)

// Enum value maps for Receipt_Status.
var (
	Receipt_Status_name = map[int32]string{
		0: "SUCCESS",
		1: "FAILED",
	}
	Receipt_Status_value = map[string]int32{
		"SUCCESS": 0,
		"FAILED":  1,
	}
)

func (x Receipt_Status) Enum() *Receipt_Status {
	p := new(Receipt_Status)
	*p = x
	return p
}

func (x Receipt_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Receipt_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_receipt_proto_enumTypes[0].Descriptor()
}

func (Receipt_Status) Type() protoreflect.EnumType {
	return &file_receipt_proto_enumTypes[0]
}

func (x Receipt_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Receipt_Status.Descriptor instead.
func (Receipt_Status) EnumDescriptor() ([]byte, []int) {
	return file_receipt_proto_rawDescGZIP(), []int{0, 0}
}

type Event_EventType int32

const (
	Event_OTHER Event_EventType = 0
)

// Enum value maps for Event_EventType.
var (
	Event_EventType_name = map[int32]string{
		0: "OTHER",
	}
	Event_EventType_value = map[string]int32{
		"OTHER": 0,
	}
)

func (x Event_EventType) Enum() *Event_EventType {
	p := new(Event_EventType)
	*p = x
	return p
}

func (x Event_EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Event_EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_receipt_proto_enumTypes[1].Descriptor()
}

func (Event_EventType) Type() protoreflect.EnumType {
	return &file_receipt_proto_enumTypes[1]
}

func (x Event_EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Event_EventType.Descriptor instead.
func (Event_EventType) EnumDescriptor() ([]byte, []int) {
	return file_receipt_proto_rawDescGZIP(), []int{2, 0}
}

type Receipt struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxHash          []byte         `protobuf:"bytes,1,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	Ret             []byte         `protobuf:"bytes,2,opt,name=ret,proto3" json:"ret,omitempty"`
	Status          Receipt_Status `protobuf:"varint,3,opt,name=status,proto3,enum=pb.Receipt_Status" json:"status,omitempty"`
	Events          []*Event       `protobuf:"bytes,4,rep,name=events,proto3" json:"events,omitempty"`
	GasUsed         uint64         `protobuf:"varint,5,opt,name=gas_used,json=gasUsed,proto3" json:"gas_used,omitempty"`
	EvmLogs         []*EvmLog      `protobuf:"bytes,6,rep,name=evm_logs,json=evmLogs,proto3" json:"evm_logs,omitempty"`
	Bloom           []byte         `protobuf:"bytes,7,opt,name=bloom,proto3" json:"bloom,omitempty"`
	ContractAddress []byte         `protobuf:"bytes,8,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
}

func (x *Receipt) Reset() {
	*x = Receipt{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipt_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Receipt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Receipt) ProtoMessage() {}

func (x *Receipt) ProtoReflect() protoreflect.Message {
	mi := &file_receipt_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Receipt.ProtoReflect.Descriptor instead.
func (*Receipt) Descriptor() ([]byte, []int) {
	return file_receipt_proto_rawDescGZIP(), []int{0}
}

func (x *Receipt) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

func (x *Receipt) GetRet() []byte {
	if x != nil {
		return x.Ret
	}
	return nil
}

func (x *Receipt) GetStatus() Receipt_Status {
	if x != nil {
		return x.Status
	}
	return Receipt_SUCCESS
}

func (x *Receipt) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

func (x *Receipt) GetGasUsed() uint64 {
	if x != nil {
		return x.GasUsed
	}
	return 0
}

func (x *Receipt) GetEvmLogs() []*EvmLog {
	if x != nil {
		return x.EvmLogs
	}
	return nil
}

func (x *Receipt) GetBloom() []byte {
	if x != nil {
		return x.Bloom
	}
	return nil
}

func (x *Receipt) GetContractAddress() []byte {
	if x != nil {
		return x.ContractAddress
	}
	return nil
}

type Receipts struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Receipts []*Receipt `protobuf:"bytes,1,rep,name=receipts,proto3" json:"receipts,omitempty"`
}

func (x *Receipts) Reset() {
	*x = Receipts{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipt_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Receipts) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Receipts) ProtoMessage() {}

func (x *Receipts) ProtoReflect() protoreflect.Message {
	mi := &file_receipt_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Receipts.ProtoReflect.Descriptor instead.
func (*Receipts) Descriptor() ([]byte, []int) {
	return file_receipt_proto_rawDescGZIP(), []int{1}
}

func (x *Receipts) GetReceipts() []*Receipt {
	if x != nil {
		return x.Receipts
	}
	return nil
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TxHash    []byte          `protobuf:"bytes,1,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	Data      []byte          `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	EventType Event_EventType `protobuf:"varint,3,opt,name=event_type,json=eventType,proto3,enum=pb.Event_EventType" json:"event_type,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipt_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_receipt_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_receipt_proto_rawDescGZIP(), []int{2}
}

func (x *Event) GetTxHash() []byte {
	if x != nil {
		return x.TxHash
	}
	return nil
}

func (x *Event) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Event) GetEventType() Event_EventType {
	if x != nil {
		return x.EventType
	}
	return Event_OTHER
}

type EvmLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address          []byte   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Topics           [][]byte `protobuf:"bytes,2,rep,name=topics,proto3" json:"topics,omitempty"`
	Data             []byte   `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	BlockNumber      uint64   `protobuf:"varint,4,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	TransactionHash  []byte   `protobuf:"bytes,5,opt,name=transaction_hash,json=transactionHash,proto3" json:"transaction_hash,omitempty"`
	TransactionIndex uint64   `protobuf:"varint,6,opt,name=transaction_index,json=transactionIndex,proto3" json:"transaction_index,omitempty"`
	BlockHash        []byte   `protobuf:"bytes,7,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
	LogIndex         uint64   `protobuf:"varint,8,opt,name=log_index,json=logIndex,proto3" json:"log_index,omitempty"`
	Removed          bool     `protobuf:"varint,9,opt,name=removed,proto3" json:"removed,omitempty"`
}

func (x *EvmLog) Reset() {
	*x = EvmLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_receipt_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EvmLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EvmLog) ProtoMessage() {}

func (x *EvmLog) ProtoReflect() protoreflect.Message {
	mi := &file_receipt_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EvmLog.ProtoReflect.Descriptor instead.
func (*EvmLog) Descriptor() ([]byte, []int) {
	return file_receipt_proto_rawDescGZIP(), []int{3}
}

func (x *EvmLog) GetAddress() []byte {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *EvmLog) GetTopics() [][]byte {
	if x != nil {
		return x.Topics
	}
	return nil
}

func (x *EvmLog) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *EvmLog) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *EvmLog) GetTransactionHash() []byte {
	if x != nil {
		return x.TransactionHash
	}
	return nil
}

func (x *EvmLog) GetTransactionIndex() uint64 {
	if x != nil {
		return x.TransactionIndex
	}
	return 0
}

func (x *EvmLog) GetBlockHash() []byte {
	if x != nil {
		return x.BlockHash
	}
	return nil
}

func (x *EvmLog) GetLogIndex() uint64 {
	if x != nil {
		return x.LogIndex
	}
	return 0
}

func (x *EvmLog) GetRemoved() bool {
	if x != nil {
		return x.Removed
	}
	return false
}

var File_receipt_proto protoreflect.FileDescriptor

var file_receipt_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x70, 0x62, 0x22, 0xa9, 0x02, 0x0a, 0x07, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x12,
	0x17, 0x0a, 0x07, 0x74, 0x78, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x72, 0x65, 0x74, 0x12, 0x2a, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x70, 0x62, 0x2e,
	0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x21, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x61, 0x73,
	0x5f, 0x75, 0x73, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x67, 0x61, 0x73,
	0x55, 0x73, 0x65, 0x64, 0x12, 0x25, 0x0a, 0x08, 0x65, 0x76, 0x6d, 0x5f, 0x6c, 0x6f, 0x67, 0x73,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x6d, 0x4c,
	0x6f, 0x67, 0x52, 0x07, 0x65, 0x76, 0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x62,
	0x6c, 0x6f, 0x6f, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x62, 0x6c, 0x6f, 0x6f,
	0x6d, 0x12, 0x29, 0x0a, 0x10, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x5f, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x21, 0x0a, 0x06,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53,
	0x53, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x22,
	0x33, 0x0a, 0x08, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x12, 0x27, 0x0a, 0x08, 0x72,
	0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e,
	0x70, 0x62, 0x2e, 0x52, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65,
	0x69, 0x70, 0x74, 0x73, 0x22, 0x80, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x17,
	0x0a, 0x07, 0x74, 0x78, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x32, 0x0a, 0x0a, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x13, 0x2e, 0x70, 0x62, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x22,
	0x16, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05,
	0x4f, 0x54, 0x48, 0x45, 0x52, 0x10, 0x00, 0x22, 0x9f, 0x02, 0x0a, 0x06, 0x45, 0x76, 0x6d, 0x4c,
	0x6f, 0x67, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a, 0x06,
	0x74, 0x6f, 0x70, 0x69, 0x63, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x10, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x48, 0x61, 0x73, 0x68, 0x12, 0x2b, 0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x10, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x68, 0x61, 0x73,
	0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61,
	0x73, 0x68, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x6f, 0x67, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x6c, 0x6f, 0x67, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12,
	0x18, 0x0a, 0x07, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2e, 0x2f,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_receipt_proto_rawDescOnce sync.Once
	file_receipt_proto_rawDescData = file_receipt_proto_rawDesc
)

func file_receipt_proto_rawDescGZIP() []byte {
	file_receipt_proto_rawDescOnce.Do(func() {
		file_receipt_proto_rawDescData = protoimpl.X.CompressGZIP(file_receipt_proto_rawDescData)
	})
	return file_receipt_proto_rawDescData
}

var file_receipt_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_receipt_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_receipt_proto_goTypes = []interface{}{
	(Receipt_Status)(0),  // 0: pb.Receipt.Status
	(Event_EventType)(0), // 1: pb.Event.EventType
	(*Receipt)(nil),      // 2: pb.Receipt
	(*Receipts)(nil),     // 3: pb.Receipts
	(*Event)(nil),        // 4: pb.Event
	(*EvmLog)(nil),       // 5: pb.EvmLog
}
var file_receipt_proto_depIdxs = []int32{
	0, // 0: pb.Receipt.status:type_name -> pb.Receipt.Status
	4, // 1: pb.Receipt.events:type_name -> pb.Event
	5, // 2: pb.Receipt.evm_logs:type_name -> pb.EvmLog
	2, // 3: pb.Receipts.receipts:type_name -> pb.Receipt
	1, // 4: pb.Event.event_type:type_name -> pb.Event.EventType
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_receipt_proto_init() }
func file_receipt_proto_init() {
	if File_receipt_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_receipt_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Receipt); i {
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
		file_receipt_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Receipts); i {
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
		file_receipt_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
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
		file_receipt_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EvmLog); i {
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
			RawDescriptor: file_receipt_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_receipt_proto_goTypes,
		DependencyIndexes: file_receipt_proto_depIdxs,
		EnumInfos:         file_receipt_proto_enumTypes,
		MessageInfos:      file_receipt_proto_msgTypes,
	}.Build()
	File_receipt_proto = out.File
	file_receipt_proto_rawDesc = nil
	file_receipt_proto_goTypes = nil
	file_receipt_proto_depIdxs = nil
}
