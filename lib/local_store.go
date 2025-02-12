package lib

import "sync"

type TransactionID int

var nextTID = 0
var newTIDMutex sync.Mutex




type Tx interface {
	Commit (tx *Tx)
}

// RWTx represents a read-write transaction.
type RWTx struct {
	// Implementation details of RWTx.
	ID TransactionID
}

func (tx *RWTx) Commit() {
}

// ROTx represents a read-only transaction.
type ROTx struct {
	// Implementation details of ROTx.
	ID TransactionID
}

func (tx *ROTx) Commit() {
}


// LocalStore represents a local transaction-capable store.
type LocalStore interface {
	NewTransaction() RWTx
	NewReadOnlyTransaction() ROTx
	Flush()
}

type FakeLocalStore struct {}

func NewFakeLocalStore() FakeLocalStore {
	return FakeLocalStore{}
}

func (ls FakeLocalStore) NewTransaction() RWTx {
	return RWTx{ ID: newTID() }
}

func (ls FakeLocalStore) NewReadOnlyTransaction() ROTx {
	return ROTx{ ID: newTID() }
}

func (ls FakeLocalStore) Flush() {}



func newTID() TransactionID {
	newTIDMutex.Lock()
	defer newTIDMutex.Unlock()
	curID := nextTID
	nextTID++
	return TransactionID(curID)
}
