package lib

import (
	_ "context"
)

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}



type ProposeArgs[EntryType any] struct {
	//Context context.Context
	E EntryType
}

type ProposeReply[ReturnType any] struct {
	Result Future[ReturnType]
}

type SyncArgs struct {
	//Context context.Context
}

type SyncReply struct {
	Result Future[ROTx]
}

// copied from kvstore -- how to not break this abstraction?
type KV struct {
	Key string
	Val string
}


//Sync(ctx context.Context) Future[ROTx]
//RegisterUpcall(app *IApplicator[ReturnType, EntryType])
//SetTrimPrefix(pos LogPos)


