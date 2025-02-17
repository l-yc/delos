package main

import (
	"context"
	"encoding/gob"
	"flag"
	"log"
	"net"
	"net/rpc"
	"sync"

	"go.etcd.io/raft/v3/raftpb"

	. "github.com/l-yc/delos/lib"
)



type RPCEngine struct {
	client *rpc.Client
}

func (self RPCEngine) Propose(ctx context.Context, e Entry) Future[string] {	
	// Synchronous call
	//args := &ProposeArgs[Entry]{ Context: ctx, E: e }
	args := &ProposeArgs[Entry]{ E: e }

	var reply ProposeReply[string]
	err := self.client.Call("API.Propose", args, &reply)
	if err != nil {
		log.Fatal("API error:", err)
	}
	log.Println(reply)
	//log.Printf("Arith: %d*%d=%d", args.A, args.B, reply)

	return reply.Result
}

// Sync synchronizes the state with the shared log tail.
func (self RPCEngine) Sync(ctx context.Context) Future[ROTx] {

	//args := &SyncArgs{ Context: ctx }
	args := &SyncArgs{}

	log.Println("calling sync with", args)
	var reply SyncReply
	err := self.client.Call("API.Sync", args, &reply)
	if err != nil {
		log.Fatal("API error:", err)
	}
	log.Println(reply)
	//log.Printf("Arith: %d*%d=%d", args.A, args.B, reply)

	return reply.Result
}

func (self RPCEngine) RegisterUpcall(app *IApplicator[string, Entry]) {

	args := &RegisterUpcallArgs{ Addr: "localhost:42585" }

	log.Println("calling registerupcall with", args)
	var reply RegisterUpcallReply
	err := self.client.Call("API.RegisterUpcall", args, &reply)
	if err != nil {
		log.Fatal("API error:", err)
	}
	log.Println(reply)
	//log.Printf("Arith: %d*%d=%d", args.A, args.B, reply)
}

// SetTrimPrefix sets the trim prefix for garbage collection.
func (self RPCEngine) SetTrimPrefix(pos LogPos) {
}

//type IEngine[ReturnType any, EntryType any] interface {
//	Propose(ctx context.Context, e EntryType) Future[ReturnType]
//	Sync(ctx context.Context) Future[ROTx]
//	RegisterUpcall(app *IApplicator[ReturnType, EntryType])
//	SetTrimPrefix(pos LogPos)
//}






var kvs *KVStore

type API int

func (a *API) Apply(args *ApplyArgs[Entry], reply *ApplyReply[string]) error {
	reply.Result = kvs.Apply(args.Txn, args.E, args.Pos)
	return nil
}







func main() {
	gob.Register(context.Background())
	gob.Register(KV{})

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()

		addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:42585")
		if err != nil {
			log.Fatal(err)
		}

		inbound, err := net.ListenTCP("tcp", addy)
		if err != nil {
			log.Fatal(err)
		}

		api := new(API)
		rpc.Register(api)
		rpc.Accept(inbound)
	}()

	client, err := rpc.Dial("tcp", "localhost:42586")
	if err != nil {
		log.Fatal("could not connect to rpc error:", err)
	}
	engine := RPCEngine{
		client,
	}
	log.Println("connected to rpc")













	kvport := flag.Int("port", 1337, "key-value server port")
	flag.Parse()


	//sharedLog := NewSimpleVirtualLog()
	////localStore := NewFakeLocalStore()
	//var localStore LocalStore = NewFakeLocalStore()
	//be := NewBaseEngine(&sharedLog, &localStore)
	//log.Println("created base engine", be)


	//var test2 IEngine[string, Entry] = be
	var test2 IEngine[string, Entry] = engine
	kvs = NewKVStore(&test2)


	//go func() {
	//	log.Println("set in kv") 
	//	kvs.ProposeSet("a", "1")
	//	log.Println("set in kv done") 
	//	//kvs.Set("b", "2")
	//	//log.Println("finish set in kv") 
	//}()

	//for {}
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)
	errorC := make(chan error)
	defer close(confChangeC)

	// taken directly from https://github.com/etcd-io/etcd/blob/main/contrib/raftexample/httpapi.go
	log.Println("serving http.....")
	serveHTTPKVAPI(kvs, *kvport, confChangeC, errorC)

	wg.Wait()
}
