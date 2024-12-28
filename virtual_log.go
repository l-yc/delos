package main

import (
	"go.etcd.io/raft/v3"

	"fmt"
)

// MARK: Types {{{

// LogPos represents a position in the log.
type LogPos int

// Entry represents an entry in the log.
type Entry struct {
	Payload string // Example payload field; can be replaced with a more complex type.
}

// LogCfg represents a configuration for reconfiguration operations.
type LogCfg struct {
	// Fields to define the log configuration.
	ConfigID int
}

/// }}}

// MARK: Loglet API {{{

// class ILoglet {
// 	logpos_t append(Entry payload);
// 	pair<logpos_t,bool> checkTail();
// 	Entry readNext(logpos_t min, logpos_t max);
// 	logpos_t prefixTrim(logpos_t trimpos);
// 	void seal();
// }

type ILoglet interface {
	Append(payload Entry) LogPos
	CheckTail() (LogPos, bool)
	ReadNext(min LogPos, max LogPos) Entry
	PrefixTrim(trimPos LogPos) LogPos
	Seal()
}

// class IVirtualLog : public ILoglet {
// 	bool reconfigExtend(LogCfg newcfg);
// 	bool reconfigTruncate();
// 	bool reconfigModify(LogCfg newcfg);
// }

type IVirtualLog interface {
	ILoglet
	ReconfigExtend(newCfg LogCfg) bool
	ReconfigTruncate() bool
	ReconfigModify(newCfg LogCfg) bool
}

/// }}}

type VirtualLog interface {
	Append(entry Entry) LogPos
	CheckTail() LogPos
	Read(min LogPos, max LogPos) []Entry
	Trim(prefix LogPos)
}



type FakeVirtualLog struct {
	n raft.Node
}

func NewFakeVirtualLog() FakeVirtualLog {
	storage := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	// Set peer list to all nodes in the cluster.
	// Note that other nodes in the cluster, apart from this node, need to be started separately.
	//n := raft.StartNode(c, []raft.Peer{{ID: 0x01}, {ID: 0x02}, {ID: 0x03}})


	// Create storage and config as shown above.
	// Set peer list to itself, so this node can become the leader of this single-node cluster.
	peers := []raft.Peer{{ID: 0x01}}
	n := raft.StartNode(c, peers)

	return FakeVirtualLog{
		n,
	}
}

func (vl FakeVirtualLog) Run() {
	fmt.Println(vl.n);

	//for {
	//	select {
	//	case <-s.Ticker:
	//		n.Tick()
	//	case rd := <-s.Node.Ready():
	//		saveToStorage(rd.HardState, rd.Entries, rd.Snapshot)
	//		send(rd.Messages)
	//		if !raft.IsEmptySnap(rd.Snapshot) {
	//			processSnapshot(rd.Snapshot)
	//		}
	//		for _, entry := range rd.CommittedEntries {
	//			process(entry)
	//			if entry.Type == raftpb.EntryConfChange {
	//				var cc raftpb.ConfChange
	//				cc.Unmarshal(entry.Data)
	//				s.Node.ApplyConfChange(cc)
	//			}
	//		}
	//		s.Node.Advance()
	//	case <-s.done:
	//		return
	//	}
	//}
}



func (vl FakeVirtualLog) Append(entry Entry) LogPos {
	// vl.n.Propose(ctx, data)
	return -1
}

func (vl FakeVirtualLog) CheckTail() LogPos {
	return -1
}

func (vl FakeVirtualLog) Read(min LogPos, max LogPos) []Entry {
	return []Entry{}
}

func (vl FakeVirtualLog) Trim(prefix LogPos) {

}
