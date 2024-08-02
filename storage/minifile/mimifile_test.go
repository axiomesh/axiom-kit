package minifile

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBatchFile(t *testing.T) {
	path := t.TempDir()

	b, err := New(path)
	assert.Nil(t, err)
	assert.Equal(t, path, b.path)

	err = b.Close()
	assert.Nil(t, err)

	b, err = New("")
	assert.Nil(t, err)

	err = b.Close()
	assert.Nil(t, err)

	b, err = New(".")
	assert.Nil(t, err)
	assert.NotEqual(t, ".", b.path)
	path = b.path
	defer func() {
		err = os.Remove(filepath.Join(path, FLOCK))
		assert.Nil(t, err)
	}()

	b, err = New(".")
	assert.NotNil(t, err)

}

func TestBatchFile_Put(t *testing.T) {
	path := t.TempDir()

	b, err := New(path)
	assert.Nil(t, err)

	key := "abc"
	val := []byte{1, 2, 3}

	err = b.Put(key, val)
	assert.Nil(t, err)

	v, e := b.Get(key)
	assert.Nil(t, e)
	assert.Equal(t, val, v)

	e = b.Delete(key)
	assert.Nil(t, e)

	v, e = b.Get(key)
	assert.Nil(t, e)
	assert.Nil(t, v)

	key = "abc"
	val = []byte{}

	err = b.Put(key, val)
	assert.Nil(t, err)

	v, e = b.Get(key)
	assert.Nil(t, e)
	assert.Equal(t, val, v)
}

func TestBatchFile_Prefix(t *testing.T) {
	path := t.TempDir()

	b, err := New(path)
	assert.Nil(t, err)

	prefix := "abc"
	val := []byte{1, 2, 3}

	err = b.Put(prefix, val)
	assert.Nil(t, err)

	err = b.Put(prefix+"1", val)
	assert.Nil(t, err)

	err = b.Put(prefix+"2", val)
	assert.Nil(t, err)

	err = b.Put("2", val)
	assert.Nil(t, err)

	files, err := b.GetAll("")
	assert.Nil(t, err)
	assert.Equal(t, 4, len(files))

	err = b.DeleteAll("")
	assert.Nil(t, err)

	m, err := b.GetAll("")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(m))

	err = b.Close()
	assert.Nil(t, err)
}

func BenchmarkMiniFile_Get(b *testing.B) {
	path := b.TempDir()

	f, err := New(path)
	assert.Nil(b, err)

	val := make([]byte, 1024*1024*1)
	for k := 0; k < len(val); k++ {
		val[k] = byte(rand.Int63n(128))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("abc%d.%d", i, j)
			err = f.Put(key, val)
			assert.Nil(b, err)

			v, e := f.Get(key)
			assert.Nil(b, e)
			assert.Equal(b, val, v)
		}

		e := f.DeleteAll("")
		assert.Nil(b, e)
	}

	err = f.Close()
	assert.Nil(b, err)
}

func TestMiniFile_AscendRange(t *testing.T) {
	path := t.TempDir()

	b, err := New(path)
	assert.Nil(t, err)

	data := []struct {
		key   string
		value []byte
	}{
		{"alice", []byte("alice_val")},
		{"bob", []byte("bob_val")},
		{"carol", []byte("carol_val")},
		{"dave", []byte("dave_val")},
		{"eric", []byte("eric_val")},
		{"fred", []byte("fred_val")},
		{"george", []byte("george_val")},
		{"harry", []byte("harry_val")},
	}

	for _, d := range data {
		err = b.Put(d.key, d.value)
		assert.Nil(t, err)
	}

	// Test AscendRange
	var resultAscendRange []string

	b.AscendRange([]byte("dave"), []byte("george"), func(key, value []byte) (bool, error) {
		resultAscendRange = append(resultAscendRange, string(key))
		return true, nil
	})
	assert.Equal(t, []string{"dave", "eric", "fred"}, resultAscendRange)
}

func TestMiniFile_DescendRange(t *testing.T) {
	path := t.TempDir()

	b, err := New(path)
	assert.Nil(t, err)

	data := []struct {
		key   string
		value []byte
	}{
		{"alice", []byte("alice_val")},
		{"bob", []byte("bob_val")},
		{"carol", []byte("carol_val")},
		{"dave", []byte("dave_val")},
		{"eric", []byte("eric_val")},
		{"fred", []byte("fred_val")},
		{"george", []byte("george_val")},
		{"harry", []byte("harry_val")},
	}

	for _, d := range data {
		err = b.Put(d.key, d.value)
		assert.Nil(t, err)
	}

	// Test AscendRange
	var resultAscendRange []string

	b.DescendRange([]byte("harry"), []byte("bob"), func(key, value []byte) (bool, error) {
		resultAscendRange = append(resultAscendRange, string(key))
		return true, nil
	})
	assert.Equal(t, []string{"harry", "george", "fred", "eric", "dave", "carol"}, resultAscendRange)
}
