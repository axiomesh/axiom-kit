package jmt

import (
	"errors"
	"github.com/axiomesh/axiom-kit/storage"
	"github.com/axiomesh/axiom-kit/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

var (
	ErrorNotFound = errors.New("not found in DB")
)

var placeHolder = (&types.LeafNode{}).GetHash()

type JMT struct {
	root        types.Node
	rootNodeKey *types.NodeKey
	typ         []byte

	backend    storage.Storage
	pruneCache PruneCache
	trieCache  TrieCache
	dirtySet   map[string]types.Node
	pruneSet   map[string]struct{}
	logger     logrus.FieldLogger
}

type PruneArgs struct {
	Enable  bool               // whether enable pruning or not
	Journal *types.TrieJournal // if Enable is true, jmt.Commit will set Journal
}

// New load and init jmt from kv.
// Before New, there must be a mapping <rootHash, rootNodeKey> in kv.
func New(rootHash common.Hash, backend storage.Storage, trieCache TrieCache, pruneCache PruneCache, logger logrus.FieldLogger) (*JMT, error) {
	var root types.Node
	var err error
	jmt := &JMT{
		backend:    backend,
		pruneCache: pruneCache,
		trieCache:  trieCache,
		dirtySet:   make(map[string]types.Node),
		pruneSet:   make(map[string]struct{}),
		logger:     logger,
	}
	rawRootNodeKey := backend.Get(rootHash[:])
	if rawRootNodeKey == nil {
		return nil, ErrorNotFound
	}
	jmt.rootNodeKey = types.DecodeNodeKey(rawRootNodeKey)
	// root node may be leaf node or internal node
	root, err = jmt.getNode(jmt.rootNodeKey)
	if err != nil {
		return nil, err
	}
	jmt.root = root
	jmt.typ = jmt.rootNodeKey.Type
	return jmt, nil
}

func (jmt *JMT) Root() types.Node {
	return jmt.root
}

// Get finds the value according to  key in tree.
// If key isn't exist in tree, return nil with no error.
func (jmt *JMT) Get(key []byte) ([]byte, error) {
	return jmt.get(jmt.root, key, 0)
}

func (jmt *JMT) get(root types.Node, key []byte, next int) (value []byte, err error) {
	switch n := (root).(type) {
	case *types.InternalNode:
		if n.Children[key[next]] == nil {
			return nil, nil
		}
		child := n.Children[key[next]]
		nextBlkNum := child.Version
		nextNodeKey := &types.NodeKey{
			Version: nextBlkNum,
			Path:    key[:next+1],
			Type:    jmt.typ,
		}
		var nextNode types.Node
		nextNode, err = jmt.getNode(nextNodeKey)
		if err != nil {
			return nil, err
		}
		return jmt.get(nextNode, key, next+1) // get in next layer
	case *types.LeafNode:
		if !slices.Equal(n.Key, key) {
			return nil, nil
		}
		return n.Val, nil
	default:
		return nil, nil
	}
}

func (jmt *JMT) Update(version uint64, key, value []byte) error {
	if len(value) != 0 {
		newRoot, newRootNodeKey, _, err := jmt.insert(jmt.root, jmt.rootNodeKey, version, key, value, 0)
		if err != nil {
			return err
		}
		jmt.root = newRoot
		jmt.rootNodeKey = newRootNodeKey
		return nil
	}
	newRoot, newRootNodeKey, del, err := jmt.delete(jmt.root, jmt.rootNodeKey, version, key, 0)
	if err != nil {
		return err
	}
	if del {
		jmt.root = newRoot
		if newRoot == nil {
			jmt.rootNodeKey = &types.NodeKey{
				Version: version,
				Path:    []byte{},
				Type:    jmt.typ,
			}
		} else {
			jmt.rootNodeKey = newRootNodeKey
		}
	}
	return nil
}

// insert generate new node or update old node.
// insert will generate a new tree, and return its root.
// nodes in the updated path will be traced in pruneCache to be flushed later.
// parameter "next" means the position tha has not been addressed in current node.
func (jmt *JMT) insert(currentNode types.Node, currentNodeKey *types.NodeKey, version uint64, key, value []byte, next int) (newRoot types.Node, newRootNodeKey *types.NodeKey, isLeaf bool, err error) {
	switch n := (currentNode).(type) {
	case nil:
		// empty tree, then generate a new LeafNode
		newLeaf := &types.LeafNode{
			Key: key,
			Val: value,
		}
		newLeaf.Hash = newLeaf.GetHash()
		nk := jmt.traceDirtyNode(version, key[:next], newLeaf)
		nk.Type = jmt.typ
		return newLeaf, nk, true, nil
	case *types.InternalNode:
		var nextNode types.Node
		var nextNodeKey *types.NodeKey
		if n.Children[key[next]] != nil {
			// if slot isn't empty, get the child node in tha slot for addressing next layer
			child := n.Children[key[next]]
			nextBlkNum := child.Version
			nextNodeKey = &types.NodeKey{
				Version: nextBlkNum,
				Path:    key[:next+1],
				Type:    jmt.typ,
			}
			nextNode, err = jmt.getNode(nextNodeKey)
			if err != nil {
				return nil, nil, false, err
			}
		}
		// insert in subtree, and get the root of new subtree
		newChildNode, _, leaf, err := jmt.insert(nextNode, nextNodeKey, version, key, value, next+1)
		if err != nil {
			return nil, nil, false, err
		}
		newInternalNode := n.Copy().(*types.InternalNode)
		newInternalNode.Children[key[next]] = &types.Child{
			Version: version,
			Hash:    newChildNode.GetHash(),
			Leaf:    leaf,
		}
		jmt.tracePruningNode(currentNodeKey)
		newInternalNodeKey := jmt.traceDirtyNode(version, key[:next], newInternalNode)
		return newInternalNode, newInternalNodeKey, false, nil
	case *types.LeafNode:
		// position before next(exclusive) is common prefix of two leaf nodes, which may need split
		// case 1: two leaf nodes have the same key, which means update
		if slices.Equal(n.Key, key) {
			jmt.tracePruningNode(currentNodeKey)
			return jmt.insert(nil, nil, version, key, value, next)
		}
		// case 2: two leaf nodes have different key, need split into a list of InternalNodes
		newLeaf := &types.LeafNode{
			Key: key,
			Val: value,
		}
		newLeaf.Hash = newLeaf.GetHash()
		newInternalNode, newInternalNodeKey := jmt.splitLeafNode(n, currentNodeKey, newLeaf, version, next)
		return newInternalNode, newInternalNodeKey, false, nil
	}
	return nil, nil, false, nil
}

// delete will find the target key from current subtree, and adjust the structure of new tree,
// then returns root node of new tree.
// parameter "next" means the position that has not been addressed in current node.
func (jmt *JMT) delete(currentNode types.Node, currentNodeKey *types.NodeKey, version uint64, key []byte, next int) (newRoot types.Node, newRootNodeKey *types.NodeKey, deleted bool, err error) {
	switch n := (currentNode).(type) {
	case *types.InternalNode:
		// case 1: delete in subtree recursively, then adjust self structure if needed to maintain sparse
		var nextNode types.Node
		var nextNodeKey *types.NodeKey
		if n.Children[key[next]] == nil {
			// target child slot is empty，no-op
			return nil, nil, false, nil
		}
		// target child slot isn't empty，then delete in this subtree recursively
		child := n.Children[key[next]]
		nextBlkNum := child.Version
		nextNodeKey = &types.NodeKey{
			Version: nextBlkNum,
			Path:    key[:next+1],
			Type:    jmt.typ,
		}
		nextNode, err = jmt.getNode(nextNodeKey)
		if err != nil {
			return nil, nil, false, err
		}
		newNextNode, newNextNodeKey, del, err := jmt.delete(nextNode, nextNodeKey, version, key, next+1)
		if err != nil || !del {
			return nil, nil, false, err
		}

		// deletion op is executed indeed in subtree
		tmpRoot := n.Copy().(*types.InternalNode)
		jmt.tracePruningNode(currentNodeKey)
		switch nn := (newNextNode).(type) {
		case nil:
			// case 1.1: target slot is empty after deletion op, check if we need to compact current internal node
			tmpRoot.Children[key[next]] = nil
			pos, needCompact := isSingleLeafSubTree(tmpRoot)
			if needCompact {
				dstChild := tmpRoot.Children[pos]
				// return the compacted leaf node
				leafNk := &types.NodeKey{
					Version: dstChild.Version,
					Path:    append(key[:next], pos), // copy slice
					Type:    jmt.typ,
				}
				leaf, err := jmt.getNode(leafNk)
				if err != nil {
					return nil, nil, false, err
				}
				// remove current node and old leaf node in pruneCache
				jmt.tracePruningNode(leafNk)
				// trace new leaf node in pruneCache
				leafNk = jmt.traceDirtyNode(version, key[:next], leaf)
				return leaf, leafNk, true, nil
			}
			// current internal node doesn't need to be compacted
			nk := jmt.traceDirtyNode(version, key[:next], tmpRoot)
			return tmpRoot, nk, true, nil
		case *types.LeafNode:
			// case 1.2: subtree becomes a leaf node after deletion op, check if we need to compact current internal node
			tmpRoot.Children[key[next]] = &types.Child{
				Version: version,
				Hash:    nn.GetHash(),
				Leaf:    true,
			}
			_, needCompact := isSingleLeafSubTree(tmpRoot)
			if needCompact {
				jmt.tracePruningNode(newNextNodeKey)
				// trace origin leaf node by new index
				newNextNodeKey = jmt.traceDirtyNode(version, key[:next], nn)
				return nn, newNextNodeKey, true, nil
			}
			// current internal node doesn't need to be compacted
			nk := jmt.traceDirtyNode(version, key[:next], tmpRoot)
			return tmpRoot, nk, true, nil
		case *types.InternalNode:
			// case 1.3：subtree's root is an internal node after deletion op, so we don't need to compact current node
			tmpRoot.Children[key[next]] = &types.Child{
				Version: version,
				Hash:    nn.GetHash(),
				Leaf:    false,
			}
			nk := jmt.traceDirtyNode(version, key[:next], tmpRoot)
			return tmpRoot, nk, true, nil
		}
		return nil, nil, false, nil
	case *types.LeafNode:
		// case 2.1： target key exists in current tree, delete it
		if slices.Equal(n.Key, key) {
			jmt.tracePruningNode(currentNodeKey)
			return nil, nil, true, nil
		}
		// case 2.2： target key doesn't exist in current tree, no-op
		return nil, nil, false, nil
	default:
		// case 3: empty subtree，no-op
		return nil, nil, false, nil
	}
}

// Commit flush dirty nodes in current tree, clear pruneCache, return root hash
func (jmt *JMT) Commit(pruneArgs *PruneArgs) (rootHash common.Hash) {
	// persist <rootHash -> rootNodeKey>
	if jmt.root == nil {
		rootHash = placeHolder
	} else {
		rootHash = jmt.root.GetHash()
	}
	if pruneArgs == nil || !pruneArgs.Enable {
		// flush dirty nodes into kv
		batch := jmt.backend.NewBatch()
		for k, v := range jmt.dirtySet {
			batch.Put([]byte(k), v.Encode())
			if jmt.root != v {
				types.RecycleTrieNode(v)
			}
		}
		batch.Put(rootHash[:], jmt.rootNodeKey.Encode())
		batch.Commit()
		// gc
		jmt.dirtySet = make(map[string]types.Node)
		jmt.pruneSet = make(map[string]struct{})
		return rootHash
	}

	dirtySet := make(map[string]types.Node)
	for k, v := range jmt.dirtySet {
		dirtySet[k] = v
	}
	pruneArgs.Journal = &types.TrieJournal{
		RootHash:    rootHash,
		RootNodeKey: jmt.rootNodeKey,
		PruneSet:    jmt.pruneSet,
		DirtySet:    dirtySet,
	}
	// gc
	jmt.dirtySet = make(map[string]types.Node)
	jmt.pruneSet = make(map[string]struct{})
	return rootHash
}

func (jmt *JMT) getNode(nk *types.NodeKey) (types.Node, error) {
	var nextNode types.Node
	var err error
	var nextRawNode []byte
	k := nk.Encode()

	// try in dirtySet first
	if dirty, ok := jmt.dirtySet[string(k)]; ok {
		return dirty, err
	}

	// try in pruneCache
	if jmt.pruneCache != nil {
		if v, ok := jmt.pruneCache.Get(jmt.rootNodeKey.Version, k); ok {
			nextNode = v
			jmt.logger.Debugf("[JMT-getNode] get from pruneCache, h=%v,k=%v", jmt.rootNodeKey.Version, k)
			return nextNode, err
		}
	}

	// try in trieCache
	if jmt.trieCache != nil {
		if v, ok := jmt.trieCache.Get(k); ok {
			nextNode, err = types.UnmarshalJMTNodeFromPb(v)
			if err != nil {
				return nil, err
			}
			jmt.logger.Debugf("[JMT-getNode] get from trieCache, h=%v,k=%v", jmt.rootNodeKey.Version, k)
			return nextNode, err
		}
	}

	// try in kv at last
	nextRawNode = jmt.backend.Get(k)
	nextNode, err = types.UnmarshalJMTNodeFromPb(nextRawNode)
	if err != nil {
		return nil, err
	}

	jmt.logger.Debugf("[JMT-getNode] get from kv, h=%v,k=%v", jmt.rootNodeKey.Version, k)

	return nextNode, err
}

// splitLeafNode splits common prefix of two leaf nodes into a series of internal nodes, and construct a tree.
func (jmt *JMT) splitLeafNode(origin *types.LeafNode, originNodeKey *types.NodeKey, newLeaf *types.LeafNode, version uint64, pos int) (newRoot types.Node, newRootNodeKey *types.NodeKey) {
	root := &types.InternalNode{}
	if newLeaf.Key[pos] == origin.Key[pos] {
		// case 1: current position is common prefix, continue split.
		newChildNode, _ := jmt.splitLeafNode(origin, originNodeKey, newLeaf, version, pos+1)
		root.Children[origin.Key[pos]] = &types.Child{
			Hash:    newChildNode.GetHash(),
			Version: version,
			Leaf:    false,
		}
	} else {
		// case 2: current position isn't common prefix, branch out.
		// branch out origin leaf node
		root.Children[origin.Key[pos]] = &types.Child{
			Hash:    origin.Hash,
			Version: version,
			Leaf:    true,
		}
		jmt.tracePruningNode(originNodeKey)
		jmt.traceDirtyNode(version, origin.Key[:pos+1], origin)
		// branch out new leaf node
		root.Children[newLeaf.Key[pos]] = &types.Child{
			Hash:    newLeaf.Hash,
			Version: version,
			Leaf:    true,
		}
		jmt.traceDirtyNode(version, newLeaf.Key[:pos+1], newLeaf)
	}
	nk := jmt.traceDirtyNode(version, newLeaf.Key[:pos], root)
	return root, nk
}

// tracePruningNode clear dirty node in memory pruneCache, then record its key.
func (jmt *JMT) tracePruningNode(nk *types.NodeKey) {
	k := string(nk.Encode())
	if _, ok := jmt.dirtySet[k]; ok {
		delete(jmt.dirtySet, k)
	} else {
		jmt.pruneSet[k] = struct{}{}
	}
}

// traceDirtyNode records new node in memory pruneCache.
func (jmt *JMT) traceDirtyNode(version uint64, path []byte, newNode types.Node) *types.NodeKey {
	nk := &types.NodeKey{
		Version: version,
		Path:    make([]byte, len(path)),
		Type:    jmt.typ,
	}
	copy(nk.Path, path)
	k := string(nk.Encode())
	jmt.dirtySet[k] = newNode
	delete(jmt.pruneSet, k)
	return nk
}

// isSingleLeafSubTree determine if current internal node was a single-leaf internal node.
func isSingleLeafSubTree(n *types.InternalNode) (byte, bool) {
	var lCnt, iCnt byte
	var i, pos byte
	for i = 0; i < 16; i++ {
		if n.Children[i] != nil {
			child := n.Children[i]
			if child.Leaf {
				pos = i
				lCnt++
			} else {
				iCnt++
			}
		}
	}
	isSingleLeaf := lCnt == 1 && iCnt == 0
	return pos, isSingleLeaf
}
