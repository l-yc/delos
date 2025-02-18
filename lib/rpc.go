package lib

import (
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}



type ProposeArgs[EntryType any] struct {
	E EntryType
}

type ProposeReply[ReturnType any] struct {
	Result Future[ReturnType]
}

type SyncArgs struct {
}

type SyncReply struct {
	Result Future[ROTx]
}








type ApplyArgs[EntryType any] struct {
	Txn RWTx
	E EntryType
	Pos LogPos
}

type ApplyReply[ReturnType any] struct {
	Result ReturnType
}

type RegisterUpcallArgs struct {
	Addr string
}

type RegisterUpcallReply struct {
}
