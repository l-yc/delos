package lib

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}



type ProposeArgs[EntryType any] interface {
	ctx context.Context
	e EntryType
}

type ProposeReply[ReturnType any] interface {
	Future[ReturnType]
}

//Sync(ctx context.Context) Future[ROTx]
//RegisterUpcall(app *IApplicator[ReturnType, EntryType])
//SetTrimPrefix(pos LogPos)


