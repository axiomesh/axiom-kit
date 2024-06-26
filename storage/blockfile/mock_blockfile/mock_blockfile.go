// Code generated by MockGen. DO NOT EDIT.
// Source: blockfile.go
//
// Generated by this command:
//
//	mockgen -destination mock_blockfile/mock_blockfile.go -package mock_blockfile -source blockfile.go -typed
//

// Package mock_blockfile is a generated GoMock package.
package mock_blockfile

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockBlockFile is a mock of BlockFile interface.
type MockBlockFile struct {
	ctrl     *gomock.Controller
	recorder *MockBlockFileMockRecorder
}

// MockBlockFileMockRecorder is the mock recorder for MockBlockFile.
type MockBlockFileMockRecorder struct {
	mock *MockBlockFile
}

// NewMockBlockFile creates a new mock instance.
func NewMockBlockFile(ctrl *gomock.Controller) *MockBlockFile {
	mock := &MockBlockFile{ctrl: ctrl}
	mock.recorder = &MockBlockFileMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlockFile) EXPECT() *MockBlockFileMockRecorder {
	return m.recorder
}

// AppendBlock mocks base method.
func (m *MockBlockFile) AppendBlock(number uint64, hash, header, extra, receipts, transactions []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AppendBlock", number, hash, header, extra, receipts, transactions)
	ret0, _ := ret[0].(error)
	return ret0
}

// AppendBlock indicates an expected call of AppendBlock.
func (mr *MockBlockFileMockRecorder) AppendBlock(number, hash, header, extra, receipts, transactions any) *MockBlockFileAppendBlockCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendBlock", reflect.TypeOf((*MockBlockFile)(nil).AppendBlock), number, hash, header, extra, receipts, transactions)
	return &MockBlockFileAppendBlockCall{Call: call}
}

// MockBlockFileAppendBlockCall wrap *gomock.Call
type MockBlockFileAppendBlockCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockFileAppendBlockCall) Return(err error) *MockBlockFileAppendBlockCall {
	c.Call = c.Call.Return(err)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockFileAppendBlockCall) Do(f func(uint64, []byte, []byte, []byte, []byte, []byte) error) *MockBlockFileAppendBlockCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockFileAppendBlockCall) DoAndReturn(f func(uint64, []byte, []byte, []byte, []byte, []byte) error) *MockBlockFileAppendBlockCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// BatchAppendBlock mocks base method.
func (m *MockBlockFile) BatchAppendBlock(number uint64, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions [][]byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchAppendBlock", number, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchAppendBlock indicates an expected call of BatchAppendBlock.
func (mr *MockBlockFileMockRecorder) BatchAppendBlock(number, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions any) *MockBlockFileBatchAppendBlockCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchAppendBlock", reflect.TypeOf((*MockBlockFile)(nil).BatchAppendBlock), number, listOfHash, listOfHeader, listOfExtra, listOfReceipts, listOfTransactions)
	return &MockBlockFileBatchAppendBlockCall{Call: call}
}

// MockBlockFileBatchAppendBlockCall wrap *gomock.Call
type MockBlockFileBatchAppendBlockCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockFileBatchAppendBlockCall) Return(err error) *MockBlockFileBatchAppendBlockCall {
	c.Call = c.Call.Return(err)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockFileBatchAppendBlockCall) Do(f func(uint64, [][]byte, [][]byte, [][]byte, [][]byte, [][]byte) error) *MockBlockFileBatchAppendBlockCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockFileBatchAppendBlockCall) DoAndReturn(f func(uint64, [][]byte, [][]byte, [][]byte, [][]byte, [][]byte) error) *MockBlockFileBatchAppendBlockCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Close mocks base method.
func (m *MockBlockFile) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockBlockFileMockRecorder) Close() *MockBlockFileCloseCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockBlockFile)(nil).Close))
	return &MockBlockFileCloseCall{Call: call}
}

// MockBlockFileCloseCall wrap *gomock.Call
type MockBlockFileCloseCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockFileCloseCall) Return(arg0 error) *MockBlockFileCloseCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockFileCloseCall) Do(f func() error) *MockBlockFileCloseCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockFileCloseCall) DoAndReturn(f func() error) *MockBlockFileCloseCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Get mocks base method.
func (m *MockBlockFile) Get(kind string, number uint64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", kind, number)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockBlockFileMockRecorder) Get(kind, number any) *MockBlockFileGetCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockBlockFile)(nil).Get), kind, number)
	return &MockBlockFileGetCall{Call: call}
}

// MockBlockFileGetCall wrap *gomock.Call
type MockBlockFileGetCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockFileGetCall) Return(arg0 []byte, arg1 error) *MockBlockFileGetCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockFileGetCall) Do(f func(string, uint64) ([]byte, error)) *MockBlockFileGetCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockFileGetCall) DoAndReturn(f func(string, uint64) ([]byte, error)) *MockBlockFileGetCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// NextBlockNumber mocks base method.
func (m *MockBlockFile) NextBlockNumber() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextBlockNumber")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// NextBlockNumber indicates an expected call of NextBlockNumber.
func (mr *MockBlockFileMockRecorder) NextBlockNumber() *MockBlockFileNextBlockNumberCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextBlockNumber", reflect.TypeOf((*MockBlockFile)(nil).NextBlockNumber))
	return &MockBlockFileNextBlockNumberCall{Call: call}
}

// MockBlockFileNextBlockNumberCall wrap *gomock.Call
type MockBlockFileNextBlockNumberCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockFileNextBlockNumberCall) Return(arg0 uint64) *MockBlockFileNextBlockNumberCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockFileNextBlockNumberCall) Do(f func() uint64) *MockBlockFileNextBlockNumberCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockFileNextBlockNumberCall) DoAndReturn(f func() uint64) *MockBlockFileNextBlockNumberCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// TruncateBlocks mocks base method.
func (m *MockBlockFile) TruncateBlocks(targetBlock uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TruncateBlocks", targetBlock)
	ret0, _ := ret[0].(error)
	return ret0
}

// TruncateBlocks indicates an expected call of TruncateBlocks.
func (mr *MockBlockFileMockRecorder) TruncateBlocks(targetBlock any) *MockBlockFileTruncateBlocksCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TruncateBlocks", reflect.TypeOf((*MockBlockFile)(nil).TruncateBlocks), targetBlock)
	return &MockBlockFileTruncateBlocksCall{Call: call}
}

// MockBlockFileTruncateBlocksCall wrap *gomock.Call
type MockBlockFileTruncateBlocksCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockFileTruncateBlocksCall) Return(arg0 error) *MockBlockFileTruncateBlocksCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockFileTruncateBlocksCall) Do(f func(uint64) error) *MockBlockFileTruncateBlocksCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockFileTruncateBlocksCall) DoAndReturn(f func(uint64) error) *MockBlockFileTruncateBlocksCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
