package jmt

import (
	"github.com/axiomesh/axiom-kit/types"
)

type PruneCache interface {
	Get(version uint64, key []byte) (types.Node, bool)

	Enable() bool
}

type TrieCache interface {
	Get(nk *types.NodeKey) (types.Node, bool)

	Has(nk *types.NodeKey) bool

	Enable() bool
}
