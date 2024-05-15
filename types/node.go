package types

import "encoding/json"

type NodeStatus uint8

// status constants, the order is important
const (
	NodeStatusDataSyncer NodeStatus = iota
	NodeStatusCandidate
	NodeStatusActive
	NodeStatusPendingInactive
	NodeStatusExited
)

type NodeMetaData struct {
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	ImageURL   string `json:"image_url"`
	WebsiteURL string `json:"website_url"`
}

type NodeInfo struct {
	// The node serial number is unique in the entire network.
	// Once allocated, it will not change.
	// It is allocated through the governance contract in a manner similar to the self-incrementing primary key.
	ID uint64 `json:"id"`

	// Use BLS12-381(Use of new consensus algorithm)
	ConsensusPubKey string `json:"consensus_pub_key"`

	// Use ed25519(Currently used in consensus and p2p)
	P2PPubKey string `json:"p2p_pub_key"`

	P2PID string `json:"p2p_id"`

	// Operator address, with permission to manage node (can update)
	OperatorAddress string `json:"operator_address"`

	// Meta data (can update)
	MetaData NodeMetaData `json:"meta_data"`

	// Node status
	Status NodeStatus `json:"status"`
}

func (n *NodeInfo) Marshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *NodeInfo) Unmarshal(raw []byte) error {
	return json.Unmarshal(raw, n)
}
