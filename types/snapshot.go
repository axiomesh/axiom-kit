package types

import (
	"fmt"
	"strings"

	"github.com/axiomesh/axiom-kit/hexutil"
	"github.com/axiomesh/axiom-kit/types/pb"
)

type SnapshotJournal struct {
	Journals []*SnapshotJournalEntry
}

type SnapshotJournalEntry struct {
	Address        *Address
	PrevAccount    *InnerAccount
	AccountChanged bool
	PrevStates     map[string][]byte
}

func (j *SnapshotJournal) Encode() ([]byte, error) {
	if j == nil {
		return nil, nil
	}

	journals := make([]*pb.SnapshotJournalEntry, len(j.Journals))
	for idx, entry := range j.Journals {
		prevAccBlob, err := entry.PrevAccount.Marshal()
		if err != nil {
			return nil, err
		}
		journals[idx] = &pb.SnapshotJournalEntry{
			Address:        entry.Address.Bytes(),
			PrevAccount:    prevAccBlob,
			AccountChanged: entry.AccountChanged,
			PrevStates:     entry.PrevStates,
		}
	}

	pbSnapshotJournal := &pb.SnapshotJournal{
		Journals: journals,
	}

	res, err := pbSnapshotJournal.MarshalVTStrict()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DecodeSnapshotJournal(data []byte) (*SnapshotJournal, error) {
	if len(data) == 0 {
		return nil, nil
	}
	helper := pb.SnapshotJournalFromVTPool()
	defer func() {
		helper.Reset()
		helper.ReturnToVTPool()
	}()
	err := helper.UnmarshalVT(data)
	if err != nil {
		return nil, err
	}

	res := &SnapshotJournal{
		Journals: make([]*SnapshotJournalEntry, len(helper.Journals)),
	}

	for idx, entry := range helper.Journals {
		prevAcc := &InnerAccount{}
		if len(entry.PrevAccount) > 0 {
			_ = prevAcc.Unmarshal(entry.PrevAccount)
		} else {
			prevAcc = nil
		}
		res.Journals[idx] = &SnapshotJournalEntry{
			Address:        NewAddress(entry.Address),
			PrevAccount:    prevAcc,
			AccountChanged: entry.AccountChanged,
			PrevStates:     entry.PrevStates,
		}
	}

	return res, nil
}

func (entry *SnapshotJournalEntry) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[SnapshotJournalEntry]: addr=%v,prevAccount=%v,changed=%v,prevStates=[",
		entry.Address.String(), entry.PrevAccount, entry.AccountChanged))
	for k, v := range entry.PrevStates {
		builder.WriteString(fmt.Sprintf("k=%v,v=%v;", hexutil.Encode([]byte(k)), hexutil.Encode(v)))
	}
	builder.WriteString("]")
	return builder.String()
}
