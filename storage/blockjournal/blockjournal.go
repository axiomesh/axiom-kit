package blockjournal

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/tsdb/fileutil"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type BlockJournal struct {
	maxJournalHeight uint64 // last journal block height
	minJournalHeight uint64 // current min valid journal block height
	startHeight      uint64 // the block height of starting record journal from

	mtInfo   *metaInfo // for persistent
	metaPath string    // the file path of metaInfo

	name        string
	rootDir     string
	maxFileSize uint32 // Max file size for data-files

	head   *os.File            // File descriptor for the data head of the table
	index  *os.File            // File description
	files  map[uint32]*os.File // open files
	headId uint32              // number of the currently active head file

	headBytes uint32 // Number of bytes written to the head file

	logger     logrus.FieldLogger
	lock       sync.RWMutex // Mutex protecting the data file descriptors
	appendLock sync.Mutex   // Mutex protect data appending race

	started bool
}

type indexEntry struct {
	filenum uint32 // stored as uint32 ( 4 bytes)
	offset  uint32 // stored as uint32 ( 4 bytes)
}

const (
	indexEntrySize = 8

	journalFileNamePrefix = "journal"
)

var (
	ErrorRollbackToHigherNumber  = errors.New("rollback to higher blockchain height")
	ErrorRollbackTooMuch         = errors.New("rollback too much block")
	ErrorRemoveJournalOutOfRange = errors.New("remove journal out of range")
	ErrorJournalOutOfOrder       = errors.New("add journal not in order")
)

var maxFilesize uint32 = 2 * 1024 * 1024 * 1024 // 2GB

// unmarshallBinary deserializes binary b into the rawIndex entry.
func (i *indexEntry) unmarshalBinary(b []byte) error {
	i.filenum = uint32(binary.BigEndian.Uint32(b[:4]))
	i.offset = binary.BigEndian.Uint32(b[4:8])
	return nil
}

// marshallBinary serializes the rawIndex entry into binary.
func (i *indexEntry) marshallBinary() []byte {
	b := make([]byte, indexEntrySize)
	binary.BigEndian.PutUint32(b[:4], i.filenum)
	binary.BigEndian.PutUint32(b[4:8], i.offset)
	return b
}

type metaInfo struct {
	MinJournalHeight uint64 `json:"minJournalHeight"` // before MinJournalHeight, block's journal should be removed; at least for external call they can not retrieve journals before MinJournalHeight
	StartHeight      uint64 `json:"startHeight"`      //the block height of starting record journal from
}

func NewBlockJournal(dir string, name string, logger logrus.FieldLogger) (*BlockJournal, error) {
	if info, err := os.Lstat(dir); !os.IsNotExist(err) {
		if info.Mode()&os.ModeSymlink != 0 {
			logger.WithField("path", dir).Error("Symbolic link is not supported")
			return nil, errors.New("symbolic link datadir is not supported")
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	lock, _, err := fileutil.Flock(filepath.Join(dir, name+".FLOCK"))
	if err != nil {
		return nil, err
	}

	idxName := fmt.Sprintf("%s.index", name)
	offsets, err := openBlockFileForAppend(filepath.Join(dir, idxName))
	if err != nil {
		_ = lock.Release()
		return nil, err
	}
	blockJournal := &BlockJournal{
		index:       offsets,
		files:       make(map[uint32]*os.File),
		name:        name,
		rootDir:     dir,
		metaPath:    path.Join(dir, name+".meta.json"),
		maxFileSize: maxFilesize,
		logger:      logger,
	}

	var mtInfo metaInfo
	_, err = os.Stat(blockJournal.metaPath)
	if err == nil { //exist
		var jsonData []byte
		jsonData, err = os.ReadFile(blockJournal.metaPath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonData, &mtInfo)
		if err != nil {
			return nil, err
		}
		blockJournal.started = true
		blockJournal.startHeight = mtInfo.StartHeight
		blockJournal.minJournalHeight = mtInfo.MinJournalHeight
	}
	blockJournal.mtInfo = &mtInfo

	if err := blockJournal.repair(); err != nil {
		_ = lock.Release()
		return nil, err
	}

	return blockJournal, nil
}

func (b *BlockJournal) GetJournalRange() (minJournalHeight, maxJournalHeight uint64) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	minJournalHeight = b.minJournalHeight
	maxJournalHeight = b.maxJournalHeight
	return
}

func (b *BlockJournal) RemoveJournalsBeforeBlock(height uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if height <= b.minJournalHeight {
		return nil
	}

	if height <= b.startHeight {
		return nil
	}

	b.minJournalHeight = height
	b.mtInfo.MinJournalHeight = height

	err := b.persistentMetaInfo()
	if err != nil {
		return err
	}

	go func() {
		b.lock.Lock()
		defer b.lock.Unlock()
		err = b.removeJournalFilesBeforeHeight(height)
		if err != nil {
			b.logger.Errorf("Remove journal files before height:%s err:%s", height, err)
		}
	}()
	return nil
}

func (b *BlockJournal) persistentMetaInfo() error {
	jsonData, err := json.Marshal(&b.mtInfo)
	if err != nil {
		return err
	}

	err = os.WriteFile(b.metaPath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (b *BlockJournal) removeJournalFilesBeforeHeight(height uint64) error {

	// get the filenum of block height
	buffer := make([]byte, indexEntrySize)
	var idx indexEntry
	if _, err := b.index.ReadAt(buffer, int64((height-1-b.startHeight)*indexEntrySize)); err != nil {
		return err
	}
	if err := idx.unmarshalBinary(buffer); err != nil {
		return err
	}

	// loop remove the journal file below filenum
	for i := idx.filenum - 1; i > 0; i-- {
		delete(b.files, i)
		removeJournalFileName := fmt.Sprintf("%s.%d.data", b.name, i)
		removePath := path.Join(b.rootDir, removeJournalFileName)
		_, err := os.Stat(removePath)
		if err != nil {
			break
		} else {
			err = os.Remove(removePath)
			if err != nil {
				return err
			}
		}

	}
	return nil

}

func (b *BlockJournal) repair() error {
	buffer := make([]byte, indexEntrySize)

	stat, err := b.index.Stat()
	if err != nil {
		return err
	}

	if remainder := stat.Size() % indexEntrySize; remainder != 0 {
		err := truncateBlockFile(b.index, stat.Size()-remainder)
		if err != nil {
			return err
		}
	}
	if stat, err = b.index.Stat(); err != nil {
		return err
	}
	offsetsSize := stat.Size()

	if offsetsSize == 0 {
		b.head, err = b.openFile(0, openBlockFileForAppend)
		b.headId = 0
		b.headBytes = 0
	} else {
		// Open the head file
		var (
			firstIndex  indexEntry
			lastIndex   indexEntry
			contentSize int64
			contentExp  int64
		)
		// Read index zero, determine what file is the earliest
		// and what item offset to use
		_, err = b.index.ReadAt(buffer, 0)
		if err != nil {
			return err
		}
		err = firstIndex.unmarshalBinary(buffer)
		if err != nil {
			return err
		}

		_, err = b.index.ReadAt(buffer, offsetsSize-indexEntrySize)
		if err != nil {
			return err
		}
		err = lastIndex.unmarshalBinary(buffer)
		if err != nil {
			return err
		}
		b.head, err = b.openFile(lastIndex.filenum, openBlockFileForAppend)
		if err != nil {
			return err
		}
		if stat, err = b.head.Stat(); err != nil {
			return err
		}
		contentSize = stat.Size()

		// Keep truncating both files until they come in sync
		contentExp = int64(lastIndex.offset)

		for contentExp != contentSize {
			b.logger.WithFields(logrus.Fields{
				"indexed": contentExp,
				"stored":  contentSize,
			}).Warn("Truncating dangling head")
			if contentExp < contentSize {
				if err := truncateBlockFile(b.head, contentExp); err != nil {
					return err
				}
				contentSize = contentExp
			}
			if contentExp > contentSize {
				b.logger.WithFields(logrus.Fields{
					"indexed": contentExp,
					"stored":  contentSize,
				}).Warn("Truncating dangling indexes")
				offsetsSize -= indexEntrySize
				_, err = b.index.ReadAt(buffer, offsetsSize-indexEntrySize)
				if err != nil {
					return err
				}
				var newLastIndex indexEntry
				err = newLastIndex.unmarshalBinary(buffer)
				if err != nil {
					return err
				}
				// We might have slipped back into an earlier head-file here
				if newLastIndex.filenum != lastIndex.filenum {
					// Release earlier opened file
					b.releaseFile(lastIndex.filenum)
					if b.head, err = b.openFile(newLastIndex.filenum, openBlockFileForAppend); err != nil {
						return err
					}
					if stat, err = b.head.Stat(); err != nil {
						// TODO, anything more we can do here?
						// A data file has gone missing...
						return err
					}
					contentSize = stat.Size()
				}
				lastIndex = newLastIndex
				contentExp = int64(lastIndex.offset)
			}
		}
		// Ensure all reparation changes have been written to disk
		if err := b.index.Sync(); err != nil {
			return err
		}
		if err := b.head.Sync(); err != nil {
			return err
		}

		if b.started {
			b.maxJournalHeight = b.startHeight + uint64(offsetsSize/indexEntrySize-1) // last indexEntry points to the end of the data file
		} else {
			b.maxJournalHeight = 0
		}

		b.headBytes = uint32(contentSize)
		b.headId = lastIndex.filenum

		// Close opened files and preopen all files
		if err := b.preopen(); err != nil {
			return err
		}
		b.logger.WithFields(logrus.Fields{
			"maxJournalHeight": b.maxJournalHeight,
			"size":             b.headBytes,
		}).Debug("Chain freezer table opened")
	}
	return nil
}

// Truncate discards any recent data above the provided threshold number.
func (b *BlockJournal) Truncate(height uint64) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	existing := atomic.LoadUint64(&b.maxJournalHeight)
	if existing <= height {
		return nil
	}

	// empty snapshot, no-op
	if b.minJournalHeight == 0 && b.maxJournalHeight == 0 {
		return nil
	}

	if b.maxJournalHeight < height {
		return ErrorRollbackToHigherNumber
	}

	if b.minJournalHeight > height && !(b.minJournalHeight == 1 && height == 0) {
		return ErrorRollbackTooMuch
	}

	if b.maxJournalHeight == height {
		return nil
	}

	b.logger.WithFields(logrus.Fields{
		"maxJournalHeight": existing,
		"limit":            height,
	}).Warn("Truncating block journal")

	if height < b.startHeight {
		return b.reset()
	}

	if b.minJournalHeight > height {
		return fmt.Errorf("Truncate out of range, MinJournalHeight:%d, HeightToTruncate:%d", b.minJournalHeight, height)
	}

	if err := truncateBlockFile(b.index, int64(height-b.startHeight+1)*indexEntrySize); err != nil {
		return err
	}
	// Calculate the new expected size of the data file and Truncate it
	buffer := make([]byte, indexEntrySize)
	if _, err := b.index.ReadAt(buffer, int64((height-b.startHeight)*indexEntrySize)); err != nil {
		return err
	}
	var expected indexEntry
	err := expected.unmarshalBinary(buffer)
	if err != nil {
		return err
	}

	// We might need to Truncate back to older files
	if expected.filenum != b.headId {
		// If already open for reading, force-reopen for writing
		b.releaseFile(expected.filenum)
		newHead, err := b.openFile(expected.filenum, openBlockFileForAppend)
		if err != nil {
			return err
		}
		// Release any files _after the current head -- both the previous head
		// and any files which may have been opened for reading
		b.releaseFilesAfter(expected.filenum, true)
		// Set back the historic head
		b.head = newHead
		atomic.StoreUint32(&b.headId, expected.filenum)
	}
	if err := truncateBlockFile(b.head, int64(expected.offset)); err != nil {
		return err
	}
	// All data files truncated, set internal counters and return
	atomic.StoreUint64(&b.maxJournalHeight, height)
	atomic.StoreUint32(&b.headBytes, expected.offset)

	return nil
}

func (b *BlockJournal) reset() error {
	b.maxJournalHeight = 0
	b.minJournalHeight = 0
	b.startHeight = 0
	b.mtInfo.StartHeight = 0
	b.mtInfo.MinJournalHeight = 0
	_ = os.Remove(b.metaPath)

	err := truncateBlockFile(b.index, 0)
	if err != nil {
		return err
	}

	b.started = false
	return b.repair()
}

func (b *BlockJournal) Retrieve(height uint64) ([]byte, error) {
	b.lock.RLock()

	if b.index == nil || b.head == nil {
		b.lock.RUnlock()
		return nil, errors.New("closed")
	}
	if atomic.LoadUint64(&b.maxJournalHeight) < height {
		b.lock.RUnlock()
		return nil, errors.New("out of bounds")
	}
	if b.startHeight > height {
		b.lock.RUnlock()
		return nil, errors.New("out of bounds")
	}
	startOffset, endOffset, filenum, err := b.getBounds(height - b.startHeight)
	if err != nil {
		b.lock.RUnlock()
		return nil, err
	}
	dataFile, exist := b.files[filenum]
	if !exist {
		b.lock.RUnlock()
		return nil, fmt.Errorf("missing data file %d", filenum)
	}
	blob := make([]byte, endOffset-startOffset)
	if _, err := dataFile.ReadAt(blob, int64(startOffset)); err != nil {
		b.lock.RUnlock()
		return nil, err
	}
	b.lock.RUnlock()

	return blob, nil
}

func (b *BlockJournal) Append(height uint64, blob []byte) error {
	b.appendLock.Lock()
	defer b.appendLock.Unlock()

	if b.started {
		if height != b.maxJournalHeight+1 {
			return ErrorJournalOutOfOrder
		}
	}

	err := b.doAppend(blob)
	if err != nil {
		return err
	}

	if !b.started {
		b.startFrom(height)
	}

	b.maxJournalHeight = height

	return nil
}

func (b *BlockJournal) startFrom(height uint64) error {
	b.startHeight = height
	b.mtInfo.StartHeight = b.startHeight
	b.mtInfo.MinJournalHeight = height
	b.started = true
	b.minJournalHeight = height
	err := b.persistentMetaInfo()
	return err
}

func (b *BlockJournal) doAppend(blob []byte) error {

	b.lock.RLock()
	if b.index == nil || b.head == nil {
		b.lock.RUnlock()
		return errors.New("closed")
	}
	bLen := uint32(len(blob))
	if b.headBytes+bLen > b.maxFileSize {
		b.lock.RUnlock()
		b.lock.Lock()
		nextID := atomic.LoadUint32(&b.headId) + 1
		// We open the next file in truncated mode -- if this file already
		// exists, we need to start over from scratch on it
		newHead, err := b.openFile(nextID, openBlockFileTruncated)
		if err != nil {
			b.lock.Unlock()
			return err
		}
		// Close old file, and reopen in RDONLY mode
		b.releaseFile(b.headId)
		_, err = b.openFile(b.headId, openBlockFileForReadOnly)
		if err != nil {
			b.lock.Unlock()
			return err
		}

		// Swap out the current head
		b.head = newHead
		atomic.StoreUint32(&b.headBytes, 0)
		atomic.StoreUint32(&b.headId, nextID)
		b.lock.Unlock()
		b.lock.RLock()
	}

	defer b.lock.RUnlock()
	if _, err := b.head.Write(blob); err != nil {
		return err
	}
	newOffset := atomic.AddUint32(&b.headBytes, bLen)
	idx := indexEntry{
		filenum: atomic.LoadUint32(&b.headId),
		offset:  newOffset,
	}

	// Write indexEntry
	_, _ = b.index.Write(idx.marshallBinary())
	err := b.head.Sync()
	if err != nil {
		return err
	}
	err = b.index.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (b *BlockJournal) getBounds(entryIndex uint64) (uint32, uint32, uint32, error) {
	buffer := make([]byte, indexEntrySize)
	var startIdx, endIdx indexEntry
	if _, err := b.index.ReadAt(buffer, int64((entryIndex)*indexEntrySize)); err != nil {
		return 0, 0, 0, err
	}
	if err := endIdx.unmarshalBinary(buffer); err != nil {
		return 0, 0, 0, err
	}
	if entryIndex >= 1 {
		if _, err := b.index.ReadAt(buffer, int64((entryIndex-1)*indexEntrySize)); err != nil {
			return 0, 0, 0, err
		}
		if err := startIdx.unmarshalBinary(buffer); err != nil {
			return 0, 0, 0, err
		}
	} else {
		// the first reading
		return 0, endIdx.offset, endIdx.filenum, nil
	}
	if startIdx.filenum != endIdx.filenum {
		return 0, endIdx.offset, endIdx.filenum, nil
	}
	return startIdx.offset, endIdx.offset, endIdx.filenum, nil
}

func (b *BlockJournal) preopen() (err error) {
	b.releaseFilesAfter(0, false)

	for i := b.headId - 1; i > 0; i-- {
		if _, err = b.openFile(i, openBlockFileForReadOnly); err != nil {
			if os.IsNotExist(err) {
				break
			} else {
				return err
			}
		}
	}
	b.head, err = b.openFile(b.headId, openBlockFileForAppend)
	return err
}

func (b *BlockJournal) openFile(num uint32, opener func(string) (*os.File, error)) (f *os.File, err error) {
	var exist bool
	if f, exist = b.files[num]; !exist {
		name := fmt.Sprintf("%s.%d.data", b.name, num)
		f, err = opener(filepath.Join(b.rootDir, name))
		if err != nil {
			return nil, err
		}
		b.files[num] = f
	}
	return f, err
}

// Close closes all opened files.
func (b *BlockJournal) Close() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	var errs []error
	if err := b.index.Close(); err != nil {
		errs = append(errs, err)
	}
	b.index = nil

	for _, f := range b.files {
		if err := f.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	b.head = nil

	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

func (b *BlockJournal) releaseFilesAfter(num uint32, remove bool) {
	for fnum, f := range b.files {
		if fnum > num {
			delete(b.files, fnum)
			f.Close()
			if remove {
				os.Remove(f.Name())
			}
		}
	}
}

func (b *BlockJournal) releaseFile(num uint32) {
	if f, exist := b.files[num]; exist {
		delete(b.files, num)
		f.Close()
	}
}

func truncateBlockFile(file *os.File, size int64) error {
	if err := file.Truncate(size); err != nil {
		return err
	}
	// Seek to end for append
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	return nil
}

func openBlockFileForAppend(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	// Seek to end for append
	if _, err = file.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}
	return file, nil
}

func openBlockFileTruncated(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func openBlockFileForReadOnly(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_RDONLY, 0644)
}
