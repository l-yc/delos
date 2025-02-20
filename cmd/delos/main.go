package main

import (
	"flag"
	"log"

	"net"
	"net/rpc"

	_ "go.etcd.io/raft/v3/raftpb"

	. "github.com/l-yc/delos/lib"
)




import "errors"

var be *BaseEngine

type API int

func (a *API) Propose(args *ProposeArgs[Entry], reply *ProposeReply[string]) error {
	reply.Result = be.Propose(args.E)
	return nil
}

func (a *API) Sync(args *SyncArgs, reply *SyncReply) error {
	log.Println("sync called")
	(*reply).Result = be.Sync()
	_ = errors.New("hi")
	return nil
}

func (a *API) RegisterUpcall(args *RegisterUpcallArgs, reply *RegisterUpcallReply) error {
	client, err := rpc.Dial("tcp", args.Addr)
	if err != nil {
		log.Fatal("could not connect to rpc error:", err)
	}
	app := RPCApplicator{
		client,
	}
	var test2 IApplicator[string, Entry] = app
	be.RegisterUpcall(&test2)

	log.Println("connected to rpc")
	return nil
}







type RPCApplicator struct {
	client *rpc.Client
}

func (self RPCApplicator) Apply(txn RWTx, e Entry, pos LogPos) string {
	// Synchronous call
	args := &ApplyArgs[Entry]{ E: e }

	var reply ApplyReply[string]
	err := self.client.Call("API.Apply", args, &reply)
	if err != nil {
		log.Fatal("API error:", err)
	}
	log.Println(reply)
	return reply.Result
}

func (self RPCApplicator) PostApply(e Entry, pos LogPos) {
}

func main() {
	sharedLog := NewSimpleVirtualLog()
	var localStore LocalStore = NewFakeLocalStore()
	be = NewBaseEngine(&sharedLog, &localStore)
	log.Println("created base engine", be)

	tsport := *flag.String("tcp-serv-port", "42586", "tcp server port")

	addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:" + tsport)
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
}
