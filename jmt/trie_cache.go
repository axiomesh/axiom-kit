package jmt

import (
	"github.com/axiomesh/axiom-kit/types"
)

type TrieCache interface {
	Get(version uint64, key []byte) ([]byte, bool)

	Update(version uint64, trieJournals types.TrieJournalBatch)
}
