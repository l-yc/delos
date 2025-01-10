package main

import (
	"go.etcd.io/raft/v3"
	"go.etcd.io/raft/v3/raftpb"

	"bytes"
	"context"
	"encoding/gob"
	"log"
	"time"
)

// MARK: Types {{{

// LogPos represents a position in the log.
type LogPos int

// Entry represents an entry in the log.
type Entry struct {
	Data interface{}
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

// MARK: Loglet Implementation {{{
type SimpleLoglet struct {
	node        raft.Node
	raftStorage *raft.MemoryStorage
	sealed      bool
	applyC	    chan<- []Entry
}


func NewSimpleLoglet(applyC chan<- []Entry) SimpleLoglet {
	raftStorage := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         raftStorage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	// Set peer list to all nodes in the cluster.
	// Note that other nodes in the cluster, apart from this node, need to be started separately.
	//n := raft.StartNode(c, []raft.Peer{{ID: 0x01}, {ID: 0x02}, {ID: 0x03}})


	// Create storage and config as shown above.
	// Set peer list to itself, so this node can become the leader of this single-node cluster.
	peers := []raft.Peer{{ID: 0x01}}
	node := raft.StartNode(c, peers)

	s := SimpleLoglet{
		node,
		raftStorage,
		false,
		applyC,
	}
	log.Println("created loglet")
	go s.run()
	return s
}


func (s SimpleLoglet) run() {
	log.Println("run loglet")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.node.Tick()
		case rd := <-s.node.Ready():
			//log.Println("received hardstate", rd.HardState)
			//log.Println("received entries", rd.Entries)
			//log.Println("received committed", rd.CommittedEntries)

			s.raftStorage.Append(rd.Entries)

	//		saveToStorage(rd.HardState, rd.Entries, rd.Snapshot)
	//		send(rd.Messages)
	//		if !raft.IsEmptySnap(rd.Snapshot) {
	//			processSnapshot(rd.Snapshot)
	//		}

			data := make([]Entry, 0, len(rd.CommittedEntries))
			for _, entry := range rd.CommittedEntries {
				switch entry.Type {
				case raftpb.EntryNormal:
					if len(entry.Data) == 0 {
						// ignore empty
						break
					}

					var entryd Entry
					dec := gob.NewDecoder(bytes.NewReader(entry.Data))
					if err := dec.Decode(&entryd); err != nil {
						log.Fatal("encode error:", err)
					}

					data = append(data, entryd)
				case raftpb.EntryConfChange:
					var cc raftpb.ConfChange
					cc.Unmarshal(entry.Data)
					s.node.ApplyConfChange(cc)
				}
			}
			if len(data) > 0 {
				log.Println("sending data", data)
				s.applyC <- data
			}
			s.node.Advance()
			log.Println("done")
	//	case <-s.done:
	//		return
		}
	}
}


func (s SimpleLoglet) Append(entry Entry) LogPos {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	if err := enc.Encode(entry); err != nil {
		log.Fatal("encode error:", err)
	}
	//b := data.Bytes()
	//log.Println("SimpleLoglet: propose", entry, b)
	//s.node.Propose(context.TODO(), b)
	s.node.Propose(context.TODO(), data.Bytes())

	return -1 // TODO doesn't return a logpos, so have to figure out how to get that
}

func (s SimpleLoglet) CheckTail() (LogPos, bool) {
	//return vl.n.raftLog.committed
	lastIdx, err := s.raftStorage.LastIndex()
	if err != nil {
		return -1, s.sealed // TODO handle error
	}
	return LogPos(lastIdx), s.sealed // uint64
}

func (s SimpleLoglet) ReadNext(lo LogPos, hi LogPos) Entry {
	entries, err := s.raftStorage.Entries(uint64(lo), uint64(hi), 1)
	if err != nil || len(entries) < 1 {
		return Entry{}
	}

	var entry Entry
	dec := gob.NewDecoder(bytes.NewReader(entries[0].Data))
	if err := dec.Decode(&entry); err != nil {
		log.Fatal("encode error:", err)
	}
	return entry
}

func (s SimpleLoglet) PrefixTrim(trimPos LogPos) LogPos {
	// TODO implement this later
	return -1
}

func (vl SimpleLoglet) Seal() {

}
/// }}}

// MARK: Simple Virtual Log Implementation {{{
type SimpleVirtualLog struct {
	loglet ILoglet
}


func NewSimpleVirtualLog(applyC chan<- []Entry) SimpleVirtualLog {
	loglet := NewSimpleLoglet(applyC)
	return SimpleVirtualLog{
		loglet,
	}
}

func (vl SimpleVirtualLog) Append(entry Entry) LogPos {
	return vl.loglet.Append(entry)
}

func (vl SimpleVirtualLog) CheckTail() LogPos {
	logPos, _ := vl.loglet.CheckTail()
	return logPos
}

func (vl SimpleVirtualLog) ReadNext(mini LogPos, maxi LogPos) Entry {
	return vl.loglet.ReadNext(mini, maxi)
}

func (vl SimpleVirtualLog) PrefixTrim(trimPos LogPos) LogPos {
	return vl.loglet.PrefixTrim(trimPos)
}

func (vl SimpleVirtualLog) Seal() {
	vl.loglet.Seal()
}

func (vl SimpleVirtualLog) ReconfigExtend(newCfg LogCfg) bool {
	return true
}

func (vl SimpleVirtualLog) ReconfigTruncate() bool {
	return true
}

func (vl SimpleVirtualLog) ReconfigModify(newCfg LogCfg) bool {
	return true
}
/// }}}
