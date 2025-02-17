package lib

import (
)

// Log-structued protocol API

// template <class ReturnType, class EntryType>
// class IEngine {
// 	Future<ReturnType> propose(EntryType e);
// 	Future<ROTx> sync();
// 	void registerUpcall(IApplicator<ReturnType, EntryType> app);
// 	void setTrimPrefix(logpos_t pos);
// }

type IEngine[ReturnType any, EntryType any] interface {
	Propose(e EntryType) Future[ReturnType]
	Sync() Future[ROTx]
	RegisterUpcall(app *IApplicator[ReturnType, EntryType])
	SetTrimPrefix(pos LogPos)
}

// template <class ReturnType, class EntryType>
// class IApplicator {
// 	ReturnType apply(RWTx txn, EntryType e, logpos_t pos);
// 	void postApply(EntryType e, logpos_t pos);
// }

type IApplicator[ReturnType any, EntryType any] interface {
	Apply(txn RWTx, e EntryType, pos LogPos) ReturnType
	PostApply(e EntryType, pos LogPos)
}
