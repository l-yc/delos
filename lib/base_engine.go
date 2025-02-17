package lib

import (
	"log"
	"sync"
	"time"
)

// SharedLog represents the shared log API used by BaseEngine.
type SharedLog = SimpleVirtualLog

// BaseEngine implements the IEngine interface.
type BaseEngine struct {
	sharedLog   *SharedLog
	localStore  *LocalStore
	cursor      LogPos // for local store
	mu          sync.Mutex
	syncQueue   []chan ROTx
	trimPrefix  LogPos
	stopGC      chan struct{}
	wg          sync.WaitGroup

	app			*IApplicator[string, Entry]
}

// NewBaseEngine creates a new BaseEngine instance.
func NewBaseEngine(sharedLog *SharedLog, localStore *LocalStore) *BaseEngine {
	be := &BaseEngine{
		sharedLog:   sharedLog,
		localStore:  localStore,
		cursor:      1,
		trimPrefix:  0,
		stopGC:      make(chan struct{}),
	}

	be.startGCThread()
	return be
}

// Propose appends an entry to the shared log and ensures it is applied.
func (be *BaseEngine) Propose(e Entry) Future[string] {
	//result := make(chan ROTx, 1)
	result := make(chan string, 1)

	go func() {
		// Append the entry to the shared log.
		pos := be.sharedLog.Append(e)

		// Play the log forward to the appended entry.
		be.playLog(pos)

		// Create a read-only transaction reflecting the committed state.
		//roTx := (*be.localStore).NewReadOnlyTransaction()
		//result <- roTx
		result <- "ok"
	}()

	return Future[string]{Result: <-result}
}

// Sync synchronizes the state with the shared log tail.
func (be *BaseEngine) Sync() Future[ROTx] {
	result := make(chan ROTx, 1)

	//be.mu.Lock()
	//be.syncQueue = append(be.syncQueue, result)
	//be.mu.Unlock()

	go func() {
		log.Println("starting sync")
		// Fetch the shared log tail.
		tail := be.sharedLog.CheckTail()
		log.Println("tail is", tail)

		// Play the log forward to the tail position.
		be.playLog(tail)
		log.Println("played log")

		// Process queued sync calls once the tail is reached.
		be.mu.Lock()
		//for _, ch := range be.syncQueue {
		//	roTx := (*be.localStore).NewReadOnlyTransaction()
		//	ch <- roTx
		//}
		//be.syncQueue = nil
		roTx := (*be.localStore).NewReadOnlyTransaction()
		result <- roTx

		log.Println("sync ok")
		be.mu.Unlock()
	}()

	return Future[ROTx]{Result: <-result}
}

func (be *BaseEngine) RegisterUpcall(app *IApplicator[string, Entry]) {
	be.app = app
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

	log.Println("playing log from", be.cursor, "to", target)
	for ; be.cursor <= target; be.cursor++ {
		entry := be.sharedLog.ReadNext(be.cursor, target + 1)

		// apply
		txn := (*be.localStore).NewTransaction()
		entry = entry
		(*be.app).Apply(txn, entry, be.cursor) // FIXME add this back after sure it won't deadlock
		txn.Commit()
	}
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
				be.sharedLog.PrefixTrim(prefix)
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
}

