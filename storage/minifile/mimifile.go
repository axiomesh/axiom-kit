package minifile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/btree"
	"github.com/prometheus/tsdb/fileutil"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const FLOCK = "FLOCK"

type MiniFile struct {
	path         string
	instanceLock fileutil.Releaser // File-system lock to prevent double opens
	lock         *sync.Mutex
	closed       int64
	keyIndex     *btree.BTree
}

type BTreeItem struct {
	Key string
}

func (item BTreeItem) Less(than btree.Item) bool {
	return item.Key < than.(BTreeItem).Key
}

func New(path string) (*MiniFile, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	flock, _, err := fileutil.Flock(filepath.Join(abs, FLOCK))
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(abs, 0755)
	if err != nil {
		return nil, err
	}
	return &MiniFile{
		path:         abs,
		instanceLock: flock,
		lock:         &sync.Mutex{},
		keyIndex:     btree.New(32),
	}, nil
}

func (mf *MiniFile) Put(key string, value []byte) error {
	if mf.isClosed() {
		return errors.New("the miniFile storage is closed")
	}

	if key == "" {
		return errors.New("store file with empty key")
	}

	crc := make([]byte, 4)
	binary.LittleEndian.PutUint32(crc, util.NewCRC(value).Value())
	value = append(value, crc...)

	mf.lock.Lock()
	defer mf.lock.Unlock()

	name := filepath.Join(mf.path, key)

	if err := os.WriteFile(name, value, 0644); err != nil {
		return fmt.Errorf("fail to write file %s: %w", name, err)
	}

	mf.keyIndex.ReplaceOrInsert(BTreeItem{Key: key})
	return nil
}

func (mf *MiniFile) Delete(key string) error {
	if mf.isClosed() {
		return errors.New("the miniFile storage is closed")
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	err := os.Remove(filepath.Join(mf.path, key))
	if err != nil && isNoFileError(err) {
		return nil
	} else {
		mf.keyIndex.Delete(BTreeItem{Key: key})
	}

	return err
}

func (mf *MiniFile) AscendRange(startKey, endKey []byte, handleFn func(k []byte, v []byte) (bool, error)) {
	if mf.isClosed() {
		return
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	start := BTreeItem{Key: string(startKey)}
	end := BTreeItem{Key: string(endKey)}

	mf.keyIndex.AscendRange(start, end, func(item btree.Item) bool {
		btreeItem := item.(BTreeItem)
		val, err := mf.get(btreeItem.Key)
		if err != nil || val == nil {
			return false
		}
		continueIteration, err := handleFn([]byte(btreeItem.Key), val)
		if err != nil || !continueIteration {
			return false
		}
		return true
	})
}

func (mf *MiniFile) DescendRange(startKey, endKey []byte, handleFn func(k []byte, v []byte) (bool, error)) {
	if mf.isClosed() {
		return
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	start := BTreeItem{Key: string(startKey)}
	end := BTreeItem{Key: string(endKey)}
	mf.keyIndex.DescendRange(start, end, func(item btree.Item) bool {
		btreeItem := item.(BTreeItem)
		val, err := mf.get(btreeItem.Key)
		if err != nil || val == nil {
			return false
		}
		continueIteration, err := handleFn([]byte(btreeItem.Key), val)
		if err != nil || !continueIteration {
			return false
		}
		return true
	})
}

func (mf *MiniFile) Get(key string) ([]byte, error) {
	if mf.isClosed() {
		return nil, errors.New("the miniFile storage is closed")
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	val, err := mf.get(key)
	if err != nil {
		_ = os.Remove(filepath.Join(mf.path, key))
		return nil, err
	}

	return val, nil
}

func (mf *MiniFile) get(key string) ([]byte, error) {
	name := filepath.Join(mf.path, key)
	val, err := os.ReadFile(name)
	if err != nil {
		if isNoFileError(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(val) < 4 {
		return nil, fmt.Errorf("file %s is corrupted", key)
	}

	crc := make([]byte, 4)
	binary.LittleEndian.PutUint32(crc, util.NewCRC(val[:len(val)-4]).Value())
	if !bytes.Equal(crc, val[len(val)-4:]) {
		return nil, fmt.Errorf("CRC checksum is not correct for %s", key)
	}

	return val[:len(val)-4], nil
}

func (mf *MiniFile) Has(key string) (bool, error) {
	val, err := mf.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (mf *MiniFile) Close() error {
	if mf.isClosed() {
		return nil
	}
	atomic.StoreInt64(&mf.closed, 1)
	return mf.instanceLock.Release()
}

func (mf *MiniFile) GetAll(prefix string) (map[string][]byte, error) {
	if mf.isClosed() {
		return nil, errors.New("the miniFile storage is closed")
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	all := make(map[string][]byte)

	files, err := mf.prefix(prefix)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		val, err := mf.get(file)
		if err != nil {
			_ = os.Remove(filepath.Join(mf.path, file))
			return nil, err
		}

		all[file] = val
	}

	return all, nil
}

func (mf *MiniFile) DeleteAll(prefix string) error {
	if mf.isClosed() {
		return errors.New("the miniFile storage is closed")
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	files, err := mf.prefix(prefix)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := os.Remove(filepath.Join(mf.path, file))
		if err != nil && !isNoFileError(err) {
			return fmt.Errorf("remove file %s failed: %w", file, err)
		}
	}

	return nil
}

func (mf *MiniFile) prefix(prefix string) ([]string, error) {
	if mf.isClosed() {
		return nil, errors.New("the miniFile storage is closed")
	}

	var files []string

	if err := filepath.Walk(mf.path, func(path string, info os.FileInfo, err error) error {
		if path != mf.path {
			name := filepath.Base(path)
			if strings.HasPrefix(name, prefix) && name != "FLOCK" {
				files = append(files, name)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func isNoFileError(err error) bool {
	return strings.Contains(err.Error(), "no such file or directory")
}

func (mf *MiniFile) isClosed() bool {
	return atomic.LoadInt64(&mf.closed) == 1
}
