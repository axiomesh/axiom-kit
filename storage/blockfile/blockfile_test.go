package blockfile

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/log"
	"github.com/axiomesh/axiom-kit/types"
)

func getChunk(size int, b int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(b)
	}
	return data
}

func getStoragePath(t *testing.T) string {
	return path.Join(t.TempDir(), "blockfile")
}

func TestBlockFileBasics(t *testing.T) {
	f, err := NewBlockFile(getStoragePath(t), log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()

	tests := []struct {
		name string
		f    BlockFile
	}{
		{
			name: "file",
			f:    f,
		},
		{
			name: "memory",
			f:    NewMemory(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.f
			err = f.TruncateBlocks(uint64(0))
			assert.Nil(t, err)
			err = f.AppendBlock(uint64(0), types.NewHash([]byte{1}).Bytes(), []byte("0"), []byte(""), []byte("0"), []byte("0"))
			assert.Nil(t, err)
			num := f.NextBlockNumber()
			assert.Equal(t, uint64(1), num)

			_, err = f.Get(BlockFileHeaderTable, uint64(0))
			assert.Nil(t, err)

			emptyExtra, err := f.Get(BlockFileExtraTable, uint64(0))
			assert.Nil(t, err)
			assert.Empty(t, emptyExtra)

			err = f.AppendBlock(uint64(1), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			assert.Nil(t, err)
			num = f.NextBlockNumber()
			assert.Equal(t, uint64(2), num)

			extra, err := f.Get(BlockFileExtraTable, uint64(1))
			assert.Nil(t, err)
			assert.EqualValues(t, []byte("1"), extra)

			err = f.TruncateBlocks(uint64(1))
			assert.Nil(t, err)
			assert.Equal(t, uint64(2), f.NextBlockNumber())

			err = f.TruncateBlocks(uint64(2))
			assert.Nil(t, err)
			assert.Equal(t, uint64(2), f.NextBlockNumber())

			err = f.TruncateBlocks(uint64(0))
			assert.Nil(t, err)
			assert.Equal(t, uint64(1), f.NextBlockNumber())

			_, err = f.Get(BlockFileHeaderTable, uint64(1))
			assert.Error(t, err)

			header0, err := f.Get(BlockFileHeaderTable, uint64(0))
			assert.Nil(t, err)
			assert.EqualValues(t, []byte("0"), header0)

			err = f.AppendBlock(uint64(1), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte("1"), []byte("1"), []byte("1"))
			assert.Nil(t, err)
			num = f.NextBlockNumber()
			assert.Equal(t, uint64(2), num)

			header1, err := f.Get(BlockFileHeaderTable, uint64(1))
			assert.Nil(t, err)
			assert.EqualValues(t, []byte("1"), header1)
		})
	}
}

func TestBlockTableBasics(t *testing.T) {
	// set cutoff at 50 bytes
	f, err := newTable(os.TempDir(),
		fmt.Sprintf("unittest-%d", rand.Uint64()), 2*1000*1000*1000, log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()
	// Write 15 bytes 255 times, results in 85 files
	for x := 0; x < 255; x++ {
		data := getChunk(15, x)
		f.Append(uint64(x), data)
	}
	for y := 0; y < 255; y++ {
		exp := getChunk(15, y)
		got, err := f.Retrieve(uint64(y))
		assert.Nil(t, err)
		if !bytes.Equal(got, exp) {
			t.Fatalf("test %d, got \n%x != \n%x", y, got, exp)
		}
	}
	// Check that we cannot read too far
	_, err = f.Retrieve(uint64(255))
	assert.ErrorContains(t, err, "out of bounds")
}

func TestBatchAppendBlock(t *testing.T) {
	ast := assert.New(t)
	// set cutoff at 50 bytes
	f, err := newTable(os.TempDir(),
		fmt.Sprintf("unittest-%d", rand.Uint64()), 2*1000*1000*1000, log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()

	// test batch append
	batchNum := 255
	var listOfBlob [][]byte
	for i := 0; i < batchNum; i++ {
		data := getChunk(15, i)
		listOfBlob = append(listOfBlob, data)
	}
	err = f.BatchAppend(uint64(0), listOfBlob)
	ast.Nil(err)
	for y := 0; y < batchNum; y++ {
		exp := getChunk(15, y)
		got, err := f.Retrieve(uint64(y))
		ast.Nil(err)
		if !bytes.Equal(got, exp) {
			t.Fatalf("test %d, got \n%x != \n%x", y, got, exp)
		}
	}
	// Check that we cannot read too far
	_, err = f.Retrieve(uint64(batchNum))
	assert.ErrorContains(t, err, "out of bounds")
}

func TestAppendBlocKCase1(t *testing.T) {
	f, err := newBlockFile(getStoragePath(t), log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()
	err = f.TruncateBlocks(uint64(0))
	assert.Nil(t, err)
	err = f.AppendBlock(uint64(0), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte(""), []byte("1"), []byte("1"))
	assert.Nil(t, err)
	f.tables[BlockFileHeaderTable].items = 3
	err = f.AppendBlock(uint64(1), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte(""), []byte("1"), []byte("1"))
	assert.NotNil(t, err)
}

func TestAppendBlocKCase2(t *testing.T) {
	f, err := newBlockFile(getStoragePath(t), log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()
	err = f.TruncateBlocks(uint64(0))
	assert.Nil(t, err)
	err = f.AppendBlock(uint64(0), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte(""), []byte("1"), []byte("1"))
	assert.Nil(t, err)
	f.tables[BlockFileHeaderTable].items = 3
	err = f.AppendBlock(uint64(1), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte(""), []byte("1"), []byte("1"))
	assert.NotNil(t, err)
}

func TestAppendBlocKCase4(t *testing.T) {
	f, err := newBlockFile(getStoragePath(t), log.NewWithModule("blockfile_test"))
	assert.Nil(t, err)
	defer f.Close()
	err = f.TruncateBlocks(uint64(0))
	assert.Nil(t, err)
	err = f.AppendBlock(uint64(0), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte(""), []byte("1"), []byte("1"))
	assert.Nil(t, err)
	f.tables[BlockFileReceiptsTable].items = 3
	err = f.AppendBlock(uint64(1), types.NewHash([]byte{1}).Bytes(), []byte("1"), []byte(""), []byte("1"), []byte("1"))
	assert.NotNil(t, err)
}

func TestBlockTableBasicsClosing(t *testing.T) {
	var (
		fname  = fmt.Sprintf("basics-close-%d", rand.Uint64())
		logger = log.NewWithModule("blockfile_test")
		f      *BlockTable
		err    error
	)
	f, err = newTable(os.TempDir(), fname, 2*1000*1000*1000, logger)
	assert.Nil(t, err)
	// Write 15 bytes 255 times, results in 85 files
	for x := 0; x < 255; x++ {
		data := getChunk(15, x)
		f.Append(uint64(x), data)
		f.Close()
		f, err = newTable(os.TempDir(), fname, 2*1000*1000*1000, logger)
		assert.Nil(t, err)
	}
	defer f.Close()

	for y := 0; y < 255; y++ {
		exp := getChunk(15, y)
		got, err := f.Retrieve(uint64(y))
		assert.Nil(t, err)
		if !bytes.Equal(got, exp) {
			t.Fatalf("test %d, got \n%x != \n%x", y, got, exp)
		}
		f.Close()
		f, err = newTable(os.TempDir(), fname, 2*1000*1000*1000, logger)
		assert.Nil(t, err)
	}
}

func TestFreezerTruncate(t *testing.T) {
	fname := fmt.Sprintf("truncation-%d", rand.Uint64())
	logger := log.NewWithModule("blockfile_test")

	{ // Fill table
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		// Write 15 bytes 30 times
		for x := 0; x < 30; x++ {
			data := getChunk(15, x)
			f.Append(uint64(x), data)
		}
		// The last item should be there
		_, err = f.Retrieve(f.items - 1)
		assert.Nil(t, err)
		f.Close()
	}
	// Reopen, truncate
	{
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		defer f.Close()
		// for x := 0; x < 20; x++ {
		// 	f.truncate(uint64(30 - x - 1)) // 150 bytes
		// }
		f.truncate(10)
		if f.items != 10 {
			t.Fatalf("expected %d items, got %d", 10, f.items)
		}
		// 45, 45, 45, 15 -- bytes should be 15
		if f.headBytes != 15 {
			t.Fatalf("expected %d bytes, got %d", 15, f.headBytes)
		}
	}
}

func TestFreezerReadAndTruncate(t *testing.T) {
	fname := fmt.Sprintf("read_truncate-%d", rand.Uint64())
	logger := log.NewWithModule("blockfile_test")
	{ // Fill table
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		// Write 15 bytes 30 times
		for x := 0; x < 30; x++ {
			data := getChunk(15, x)
			f.Append(uint64(x), data)
		}
		// The last item should be there
		_, err = f.Retrieve(f.items - 1)
		assert.Nil(t, err)
		f.Close()
	}
	// Reopen and read all files
	{
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		if f.items != 30 {
			f.Close()
			t.Fatalf("expected %d items, got %d", 0, f.items)
		}
		for y := byte(0); y < 30; y++ {
			f.Retrieve(uint64(y))
		}
		// Now, truncate back to zero
		f.truncate(0)
		// Write the data again
		for x := 0; x < 30; x++ {
			data := getChunk(15, ^x)
			err := f.Append(uint64(x), data)
			assert.Nil(t, err)
		}
		f.Close()
	}
}

func TestFreezerRepairFirstFile(t *testing.T) {
	fname := fmt.Sprintf("truncationfirst-%d", rand.Uint64())
	logger := log.NewWithModule("blockfile_test")
	{ // Fill table
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		// Write 80 bytes, splitting out into two files
		f.Append(0, getChunk(40, 0xFF))
		f.Append(1, getChunk(40, 0xEE))
		// The last item should be there
		_, err = f.Retrieve(f.items - 1)
		assert.Nil(t, err)
		f.Close()
	}
	// Truncate the file in half
	fileToCrop := filepath.Join(os.TempDir(), fmt.Sprintf("%s.0001.rdat", fname))
	{
		err := assertFileSize(fileToCrop, 40)
		assert.Nil(t, err)
		file, err := os.OpenFile(fileToCrop, os.O_RDWR, 0644)
		assert.Nil(t, err)
		file.Truncate(20)
		file.Close()
	}
	// Reopen
	{
		f, err := newTable(os.TempDir(), fname, 50, logger)
		assert.Nil(t, err)
		if f.items != 1 {
			f.Close()
			t.Fatalf("expected %d items, got %d", 0, f.items)
		}
		// Write 40 bytes
		f.Append(1, getChunk(40, 0xDD))
		f.Close()
		// Should have been truncated down to zero and then 40 written
		err = assertFileSize(fileToCrop, 40)
		assert.Nil(t, err)
	}
}

func assertFileSize(f string, size int64) error {
	stat, err := os.Stat(f)
	if err != nil {
		return err
	}
	if stat.Size() != size {
		return fmt.Errorf("error, expected size %d, got %d", size, stat.Size())
	}

	return nil
}
