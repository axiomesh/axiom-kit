package jmt

import (
	"github.com/axiomesh/axiom-kit/types"
)

type PruneCache interface {
	Get(version uint64, key []byte) (types.Node, bool)

	Enable() bool
}

type TrieCache interface {
	Get(k []byte) ([]byte, bool)
}
