package main

type Tx interface {
	Commit (tx *Tx)
}

// RWTx represents a read-write transaction.
type RWTx struct {
	// Implementation details of RWTx.
}

func (tx *RWTx) Commit() {
}

// ROTx represents a read-only transaction.
type ROTx struct {
	// Implementation details of ROTx.
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
	return RWTx{}
}

func (ls FakeLocalStore) NewReadOnlyTransaction() ROTx {
	return ROTx{}
}

func (ls FakeLocalStore) Flush() {}
