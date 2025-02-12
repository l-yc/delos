package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/rpc"

	"go.etcd.io/raft/v3/raftpb"

	. "github.com/l-yc/delos/lib"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:42586")

	// Synchronous call
	args := &Args{7,8}
	var reply int
	err = client.Call("Arith.Multiply", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	log.Printf("Arith: %d*%d=%d", args.A, args.B, reply)












	gob.Register(KV{})


	kvport := flag.Int("port", 1337, "key-value server port")
	flag.Parse()


	sharedLog := NewSimpleVirtualLog()
	//localStore := NewFakeLocalStore()
	var localStore LocalStore = NewFakeLocalStore()
	be := NewBaseEngine(&sharedLog, &localStore)
	log.Println("created base engine", be)


	var test2 IEngine[string, Entry] = be
	kvs := NewKVStore(&test2)


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
	serveHTTPKVAPI(&kvs, *kvport, confChangeC, errorC)
}
