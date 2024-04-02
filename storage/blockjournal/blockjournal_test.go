package blockjournal

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axiomesh/axiom-kit/log"
)

func getChunk(size int, b int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(b)
	}
	return data
}

func TestBlockTableBasics(t *testing.T) {
	// set cutoff at 50 bytes
	f, err := NewBlockJournal(t.TempDir(), "blockjournal_test", log.NewWithModule("blockfile_test"))
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
	assert.Equal(t, errors.New("out of bounds"), err)
}

func TestBlockTableBasics2(t *testing.T) {
	// set cutoff at 50 bytes
	f, err := NewBlockJournal(t.TempDir(), "blockjournal_test", log.NewWithModule("blockjournal_test"))
	assert.Nil(t, err)
	defer f.Close()
	// Write 15 bytes 255 times, results in 85 files
	var bytes = make([]byte, 1024*1024)
	for i := 1; i < 1000; i++ {
		rand.Read(bytes)
		f.Append(uint64(i), bytes)
	}

}

func TestFreezerTruncate(t *testing.T) {

	logger := log.NewWithModule("blockjournal_test")
	tmpDir := t.TempDir()
	{ // Fill table

		f, err := NewBlockJournal(tmpDir, "blockjournal_test", logger)
		assert.Nil(t, err)
		// Write 15 bytes 30 times
		for x := 1; x < 30; x++ {
			data := getChunk(15, x)
			f.Append(uint64(x), data)
		}
		// The last item should be there
		_, err = f.Retrieve(f.maxJournalHeight)
		assert.Nil(t, err)
		f.Close()

	}
	// Reopen, Truncate
	{
		f, err := NewBlockJournal(tmpDir, "blockjournal_test", logger)
		assert.Nil(t, err)
		defer f.Close()
		// for x := 0; x < 20; x++ {
		// 	f.Truncate(uint64(30 - x - 1)) // 150 bytes
		// }
		err = f.Truncate(10)
		assert.Nil(t, err)
		if f.maxJournalHeight != 10 {
			t.Fatalf("expected %d items, got %d", 10, f.maxJournalHeight)
		}
		if f.headBytes != 150 {
			t.Fatalf("expected %d bytes, got %d", 15, f.headBytes)
		}
	}
}

func TestFreezerReadAndTruncate(t *testing.T) {

	logger := log.NewWithModule("blockfile_test")
	tmpDir := t.TempDir()
	{ // Fill table
		f, err := NewBlockJournal(tmpDir, "blockjournal_test", logger)
		assert.Nil(t, err)
		// Write 15 bytes 30 times
		for x := 1; x < 30; x++ {
			data := getChunk(15, x)
			f.Append(uint64(x), data)
		}
		// The last item should be there
		_, err = f.Retrieve(f.maxJournalHeight)
		assert.Nil(t, err)
		f.Close()
	}
	// Reopen and read all files
	{
		f, err := NewBlockJournal(tmpDir, "blockjournal_test", logger)
		assert.Nil(t, err)
		if f.maxJournalHeight != 29 {
			f.Close()
			t.Fatalf("expected %d items, got %d", 29, int(f.maxJournalHeight))
		}
		for y := byte(1); y < 30; y++ {
			retrieve, err := f.Retrieve(uint64(y))
			assert.Nil(t, err)
			assert.Equal(t, getChunk(15, int(y)), retrieve)
		}
		// Now, Truncate back to zero
		f.Truncate(0)
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
	logger := log.NewWithModule("blockfile_test")
	tmpDir := t.TempDir()
	maxFilesize = 50
	{ // Fill table
		f, err := NewBlockJournal(tmpDir, "blockjournal_test", logger)
		assert.Nil(t, err)
		// Write 80 bytes, splitting out into two files
		f.Append(0, getChunk(40, 0xFF))
		f.Append(1, getChunk(40, 0xEE))
		// The last item should be there
		_, err = f.Retrieve(f.maxJournalHeight)
		assert.Nil(t, err)
		assert.Equal(t, uint64(1), f.maxJournalHeight)
		f.Close()
	}
	// Truncate the file in half
	fileToCrop := filepath.Join(tmpDir, fmt.Sprintf("%s.1.rdat", journalFileNamePrefix))
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
		f, err := NewBlockJournal(tmpDir, "blockjournal_test", logger)
		assert.Nil(t, err)
		if f.maxJournalHeight != 0 {
			f.Close()
			t.Fatalf("expected %d items, got %d", 0, f.maxJournalHeight)
		}
		// Write 40 bytes
		f.Append(1, getChunk(40, 0xDD))
		f.Close()
		// Should have been truncated down to zero and then 40 written
		err = assertFileSize(fileToCrop, 40)
		assert.Nil(t, err)
	}
}

func removeToHeight(t *testing.T) {

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
