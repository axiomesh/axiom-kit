package jmt

import (
	"github.com/axiomesh/axiom-kit/types"
)

type PruneCache interface {
	Get(version uint64, key []byte) (types.Node, bool)

	Update(version uint64, trieJournals types.TrieJournalBatch)
}

type TrieCache interface {
	Get(k []byte) ([]byte, bool)

	Set(k []byte, v []byte)

	Del(k []byte)
}
