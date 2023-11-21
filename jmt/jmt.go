package jmt

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/slices"

	"github.com/axiomesh/axiom-kit/storage"
)

var (
	ErrorNotFound = errors.New("not found in DB")
)

var placeHolder = LeafNode{}.hash()

type JMT struct {
	root        Node
	rootNodeKey *NodeKey
	prefix      []byte
	backend     storage.Storage
	cache       map[string]Node // todo allow concurrent r/w
	dirtyNodes  map[string]Node
}

// New load and init jmt from kv.
// Before New, there must be a mapping <rootHash, rootNodeKey> in kv.
func New(rootHash common.Hash, backend storage.Storage) (*JMT, error) {
	var root Node
	var err error
	var rootNodeKey *NodeKey
	jmt := &JMT{
		backend:    backend,
		cache:      make(map[string]Node),
		dirtyNodes: make(map[string]Node),
	}
	rawRootNodeKey := backend.Get(rootHash[:])
	if rawRootNodeKey == nil {
		return nil, ErrorNotFound
	}
	rootNodeKey = decodeNodeKey(rawRootNodeKey)
	// root node may be leaf node or internal node
	root, err = jmt.getNode(rootNodeKey)
	if err != nil {
		return nil, err
	}
	jmt.root = root
	jmt.rootNodeKey = rootNodeKey
	jmt.prefix = rootNodeKey.Prefix
	return jmt, nil
}

func (jmt *JMT) Root() Node {
	return jmt.root
}

// Traverse traverses jmt in deep-first order, and will only traverse branches whose version is GTE the specific version.
// Specially, if version is 0, Traverse will traverse the whole tree.
func (jmt *JMT) Traverse(version uint64) []*TraversedNode {
	return jmt.dfs(jmt.root, version, []byte{})
}

func (jmt *JMT) dfs(root Node, version uint64, path []byte) []*TraversedNode {
	if jmt.root == nil {
		return nil
	}
	var ret []*TraversedNode
	// traverse current sub tree
	switch n := (root).(type) {
	case InternalNode:
		var hex byte
		ret = append(ret, &TraversedNode{
			Origin: &root,
			Path:   path,
		})
		for hex = 0; hex < 16; hex++ {
			if n.Children[hex] == nil || n.Children[hex].Version < version {
				continue
			}
			child := n.Children[hex]
			nextNodeKey := &NodeKey{
				Version: child.Version,
				Path:    make([]byte, len(path)),
				Prefix:  jmt.prefix,
			}
			copy(nextNodeKey.Path, path)
			nextNodeKey.Path = append(nextNodeKey.Path, hex)

			var nextNode Node
			nextNode, err := jmt.getNode(nextNodeKey)
			if err != nil {
				return nil
			}
			subTreeNodeSet := jmt.dfs(nextNode, version, nextNodeKey.Path)
			if len(subTreeNodeSet) != 0 {
				ret = append(ret, subTreeNodeSet...)
			}
		}
		break
	case LeafNode:
		ret = append(ret, &TraversedNode{
			Origin: &root,
			Path:   path,
		})
		break
	default:
		break
	}
	return ret
}

// Get finds the value according to  key in tree.
// If key isn't exist in tree, return nil with no error.
func (jmt *JMT) Get(key []byte) ([]byte, error) {
	return jmt.get(jmt.root, key, 0)
}

func (jmt *JMT) get(root Node, key []byte, next int) (value []byte, err error) {
	switch n := (root).(type) {
	case InternalNode:
		if n.Children[key[next]] == nil {
			return nil, nil
		}
		child := n.Children[key[next]]
		nextBlkNum := child.Version
		nextNodeKey := &NodeKey{
			Version: nextBlkNum,
			Path:    key[:next+1],
			Prefix:  jmt.prefix,
		}
		var nextNode Node
		nextNode, err = jmt.getNode(nextNodeKey)
		if err != nil {
			return nil, err
		}
		return jmt.get(nextNode, key, next+1) // get in next layer
	case LeafNode:
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
			jmt.rootNodeKey = &NodeKey{
				Version: version,
				Path:    []byte{},
				Prefix:  jmt.prefix,
			}
		} else {
			jmt.rootNodeKey = newRootNodeKey
		}
	}
	return nil
}

// insert generate new node or update old node.
// insert will generate a new tree, and return its root.
// nodes in the updated path will be traced in cache to be flushed later.
// parameter "next" means the position tha has not been addressed in current node.
func (jmt *JMT) insert(currentNode Node, currentNodeKey *NodeKey, version uint64, key, value []byte, next int) (newRoot Node, newRootNodeKey *NodeKey, isLeaf bool, err error) {
	switch n := (currentNode).(type) {
	case nil:
		// empty tree, then generate a new LeafNode
		newLeaf := LeafNode{
			Key: key,
			Val: value,
		}
		newLeaf.Hash = newLeaf.hash()
		nk := jmt.traceNewNode(version, key[:next], newLeaf)
		nk.Prefix = jmt.prefix
		return newLeaf, nk, true, nil
	case InternalNode:
		var nextNode Node
		var nextNodeKey *NodeKey
		if n.Children[key[next]] != nil {
			// if slot isn't empty, get the child node in tha slot for addressing next layer
			child := n.Children[key[next]]
			nextBlkNum := child.Version
			nextNodeKey = &NodeKey{
				Version: nextBlkNum,
				Path:    key[:next+1],
				Prefix:  jmt.prefix,
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
		newInternalNode := n.copy().(InternalNode)
		newInternalNode.Children[key[next]] = &Child{
			Version: version,
			Hash:    newChildNode.hash(),
			Leaf:    leaf,
		}
		newInternalNodeKey := jmt.traceNewNode(version, key[:next], newInternalNode)
		return newInternalNode, newInternalNodeKey, false, nil
	case LeafNode:
		// position before next(exclusive) is common prefix of two leaf nodes, which may need split
		// case 1: two leaf nodes have the same key, which means update
		if slices.Equal(n.Key, key) {
			return jmt.insert(nil, nil, version, key, value, next)
		}
		// case 2: two leaf nodes have different key, need split into a list of InternalNodes
		newLeaf := LeafNode{
			Key: key,
			Val: value,
		}
		newLeaf.Hash = newLeaf.hash()
		newInternalNode, newInternalNodeKey := jmt.splitLeafNode(&n, currentNodeKey, &newLeaf, version, next)
		return newInternalNode, newInternalNodeKey, false, nil
	}
	return nil, nil, false, nil
}

// delete will find the target key from current subtree, and adjust the structure of new tree,
// then returns root node of new tree.
// parameter "next" means the position that has not been addressed in current node.
func (jmt *JMT) delete(currentNode Node, currentNodeKey *NodeKey, version uint64, key []byte, next int) (newRoot Node, newRootNodeKey *NodeKey, deleted bool, err error) {
	switch n := (currentNode).(type) {
	case InternalNode:
		// case 1: delete in subtree recursively, then adjust self structure if needed to maintain sparse
		var nextNode Node
		var nextNodeKey *NodeKey
		if n.Children[key[next]] == nil {
			// target child slot is empty，no-op
			return nil, nil, false, nil
		}
		// target child slot isn't empty，then delete in this subtree recursively
		child := n.Children[key[next]]
		nextBlkNum := child.Version
		nextNodeKey = &NodeKey{
			Version: nextBlkNum,
			Path:    key[:next+1],
			Prefix:  jmt.prefix,
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
		tmpRoot := n.copy().(InternalNode)
		switch nn := (newNextNode).(type) {
		case nil:
			// case 1.1: target slot is empty after deletion op, check if we need to compact current internal node
			tmpRoot.Children[key[next]] = nil
			pos, needCompact := isSingleLeafSubTree(&tmpRoot)
			if needCompact {
				dstChild := tmpRoot.Children[pos]
				// return the compacted leaf node
				leafNk := &NodeKey{
					Version: dstChild.Version,
					Path:    append(key[:next], pos), // copy slice
					Prefix:  jmt.prefix,
				}
				leaf, err := jmt.getNode(leafNk)
				if err != nil {
					return nil, nil, false, err
				}
				// remove current node and old leaf node in cache
				jmt.removeOldNode(currentNodeKey)
				jmt.removeOldNode(leafNk)
				// trace new leaf node in cache
				leafNk = jmt.traceNewNode(version, key[:next], leaf)
				return leaf, leafNk, true, nil
			}
			// current internal node doesn't need to be compacted
			nk := jmt.traceNewNode(version, key[:next], tmpRoot)
			return tmpRoot, nk, true, nil
		case LeafNode:
			// case 1.2: subtree becomes a leaf node after deletion op, check if we need to compact current internal node
			tmpRoot.Children[key[next]] = &Child{
				Version: version,
				Hash:    nn.hash(),
				Leaf:    true,
			}
			_, needCompact := isSingleLeafSubTree(&tmpRoot)
			if needCompact {
				// remove current internal node in cache
				jmt.removeOldNode(currentNodeKey)
				jmt.removeOldNode(newNextNodeKey)
				// trace origin leaf node by new index
				newNextNodeKey = jmt.traceNewNode(version, key[:next], nn)
				return nn, newNextNodeKey, true, nil
			}
			// current internal node doesn't need to be compacted
			nk := jmt.traceNewNode(version, key[:next], tmpRoot)
			return tmpRoot, nk, true, nil
		case InternalNode:
			// case 1.3：subtree's root is an internal node after deletion op, so we don't need to compact current node
			tmpRoot.Children[key[next]] = &Child{
				Version: version,
				Hash:    nn.hash(),
				Leaf:    false,
			}
			nk := jmt.traceNewNode(version, key[:next], tmpRoot)
			return tmpRoot, nk, true, nil
		}
		return nil, nil, false, nil
	case LeafNode:
		// case 2.1： target key exists in current tree, delete it
		if slices.Equal(n.Key, key) {
			jmt.removeOldNode(currentNodeKey)
			return nil, nil, true, nil
		}
		// case 2.2： target key doesn't exist in current tree, no-op
		return nil, nil, false, nil
	default:
		// case 3: empty subtree，no-op
		return nil, nil, false, nil
	}
}

// Commit flush dirty nodes in current tree, clear cache, return root hash
func (jmt *JMT) Commit() (rootHash common.Hash) {
	// flush dirty nodes into kv
	batch := jmt.backend.NewBatch()
	for k, v := range jmt.dirtyNodes {
		batch.Put([]byte(k), v.encode())
	}
	// persist <rootHash -> rootNodeKey>
	if jmt.root == nil {
		rootHash = placeHolder
	} else {
		rootHash = jmt.root.hash()
	}
	batch.Put(rootHash[:], jmt.rootNodeKey.encode())
	batch.Commit()
	// gc
	jmt.dirtyNodes = make(map[string]Node)
	jmt.cache = make(map[string]Node)
	return rootHash
}

func (jmt *JMT) getNode(nk *NodeKey) (Node, error) {
	var nextNode Node
	var err error
	k := nk.encode()
	if cachedNode, ok := jmt.cache[string(k)]; ok {
		// find in cache first
		nextNode = cachedNode
	} else {
		// find in kv
		nextRawNode := jmt.backend.Get(k)
		nextNode, err = decodeNode(nextRawNode)
		if err != nil {
			return nil, err
		}
		jmt.cache[string(k)] = nextNode
	}
	return nextNode, err
}

// splitLeafNode splits common prefix of two leaf nodes into a series of internal nodes, and construct a tree.
// todo maybe we can reuse origin leaf node in kv? reference diem implementation
func (jmt *JMT) splitLeafNode(origin *LeafNode, originNodeKey *NodeKey, newLeaf *LeafNode, version uint64, pos int) (newRoot Node, newRootNodeKey *NodeKey) {
	root := InternalNode{}
	if newLeaf.Key[pos] == origin.Key[pos] {
		// case 1: current position is common prefix, continue split.
		newChildNode, _ := jmt.splitLeafNode(origin, originNodeKey, newLeaf, version, pos+1)
		root.Children[origin.Key[pos]] = &Child{
			Hash:    newChildNode.hash(),
			Version: version,
			Leaf:    false,
		}
	} else {
		// case 2: current position isn't common prefix, branch out.
		// branch out origin leaf node
		root.Children[origin.Key[pos]] = &Child{
			Hash:    origin.Hash,
			Version: version,
			Leaf:    true,
		}
		jmt.removeOldNode(originNodeKey)
		jmt.traceNewNode(version, origin.Key[:pos+1], *origin)
		// branch out new leaf node
		root.Children[newLeaf.Key[pos]] = &Child{
			Hash:    newLeaf.Hash,
			Version: version,
			Leaf:    true,
		}
		jmt.traceNewNode(version, newLeaf.Key[:pos+1], *newLeaf)
	}
	nk := jmt.traceNewNode(version, newLeaf.Key[:pos], root)
	return root, nk
}

// removeOldNode clear dirty node in memory cache, not in kv.
func (jmt *JMT) removeOldNode(nk *NodeKey) {
	k := string(nk.encode())
	delete(jmt.dirtyNodes, k)
	delete(jmt.cache, k)
}

// traceNewNode records new node in memory cache.
func (jmt *JMT) traceNewNode(version uint64, path []byte, newNode Node) *NodeKey {
	nk := &NodeKey{
		Version: version,
		Path:    make([]byte, len(path)),
		Prefix:  jmt.prefix,
	}
	copy(nk.Path, path)
	k := string(nk.encode())
	jmt.dirtyNodes[k] = newNode
	jmt.cache[k] = newNode
	return nk
}

// isSingleLeafSubTree determine if current internal node was a single-leaf internal node.
func isSingleLeafSubTree(n *InternalNode) (byte, bool) {
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
