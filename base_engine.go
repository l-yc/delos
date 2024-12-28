package main

import (
	"context"
	"sync"
	"time"
)

// type IEngine[ReturnType any, EntryType any] interface {
// 	Propose(ctx context.Context, e EntryType) Future[ReturnType]
// 	Sync(ctx context.Context) Future[ROTx]
// 	RegisterUpcall(app IApplicator[ReturnType, EntryType])
// 	SetTrimPrefix(pos LogPos)
// }


// SharedLog represents the shared log API used by BaseEngine.
type SharedLog = VirtualLog

// BaseEngine implements the IEngine interface.
type BaseEngine struct {
	sharedLog   SharedLog
	localStore  LocalStore
	cursor      LogPos // for local store
	mu          sync.Mutex
	applyThread chan Entry
	syncQueue   []chan ROTx
	trimPrefix  LogPos
	stopGC      chan struct{}
	wg          sync.WaitGroup
}

// NewBaseEngine creates a new BaseEngine instance.
func NewBaseEngine(sharedLog SharedLog, localStore LocalStore) *BaseEngine {
	be := &BaseEngine{
		sharedLog:   sharedLog,
		localStore:  localStore,
		cursor:      0,
		applyThread: make(chan Entry, 100), // Buffered channel for the apply thread.
		trimPrefix:  0,
		stopGC:      make(chan struct{}),
	}
	be.startApplyThread()
	be.startGCThread()
	return be
}

// Propose appends an entry to the shared log and ensures it is applied.
func (be *BaseEngine) Propose(ctx context.Context, e Entry) Future[ROTx] {
	result := make(chan ROTx, 1)

	go func() {
		// Append the entry to the shared log.
		pos := be.sharedLog.Append(e)

		// Play the log forward to the appended entry.
		be.playLog(pos)

		// Create a read-only transaction reflecting the committed state.
		roTx := be.localStore.NewReadOnlyTransaction()
		result <- roTx
	}()

	return Future[ROTx]{Result: <-result}
}

// Sync synchronizes the state with the shared log tail.
func (be *BaseEngine) Sync(ctx context.Context) Future[ROTx] {
	result := make(chan ROTx, 1)

	be.mu.Lock()
	be.syncQueue = append(be.syncQueue, result)
	be.mu.Unlock()

	go func() {
		// Fetch the shared log tail.
		tail := be.sharedLog.CheckTail()

		// Play the log forward to the tail position.
		be.playLog(tail)

		// Process queued sync calls once the tail is reached.
		be.mu.Lock()
		for _, ch := range be.syncQueue {
			roTx := be.localStore.NewReadOnlyTransaction()
			ch <- roTx
		}
		be.syncQueue = nil
		be.mu.Unlock()
	}()

	return Future[ROTx]{Result: <-result}
}

// RegisterUpcall registers the applicator for the apply calls.
func (be *BaseEngine) RegisterUpcall(app IApplicator[any, Entry]) {
	go func() {
		for entry := range be.applyThread {
			// Create a transaction for the apply process.
			txn := be.localStore.NewTransaction()
			app.Apply(txn, entry, be.cursor)
			txn.Commit()
		}
	}()
}

// SetTrimPrefix sets the trim prefix for garbage collection.
func (be *BaseEngine) SetTrimPrefix(pos LogPos) {
	be.mu.Lock()
	be.trimPrefix = pos
	be.mu.Unlock()
}

// playLog plays the log forward from the current cursor to the target position.
func (be *BaseEngine) playLog(target LogPos) {
	be.mu.Lock()
	defer be.mu.Unlock()

	entries := be.sharedLog.Read(be.cursor, target)
	for _, entry := range entries {
		be.cursor++
		be.applyThread <- entry
	}
}

// startApplyThread spawns the apply thread.
func (be *BaseEngine) startApplyThread() {
	be.wg.Add(1)
	go func() {
		defer be.wg.Done()
		for entry := range be.applyThread {
			// The actual application logic will happen here.
			_ = entry // Process entry as needed.
		}
	}()
}

// startGCThread spawns a background garbage collection thread.
func (be *BaseEngine) startGCThread() {
	be.wg.Add(1)
	go func() {
		defer be.wg.Done()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				be.mu.Lock()
				prefix := be.trimPrefix
				be.mu.Unlock()
				be.sharedLog.Trim(prefix)
			case <-be.stopGC:
				return
			}
		}
	}()
}

// Stop stops the background threads.
func (be *BaseEngine) Stop() {
	close(be.stopGC)
	be.wg.Wait()
	close(be.applyThread)
}

