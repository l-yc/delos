package main

import (
	"context"
	"encoding/gob"
	"flag"
	"log"

	"go.etcd.io/raft/v3/raftpb"
)

func main() {
	gob.Register(KV{})


	kvport := flag.Int("port", 1337, "key-value server port")
	flag.Parse()




	applyC := make(chan []Entry)
	defer close(applyC)
	sharedLog := NewSimpleVirtualLog(applyC)
	localStore := NewFakeLocalStore()
	be := NewBaseEngine(sharedLog, localStore, applyC)
	log.Println("created base engine", be)

	proposeC := make(chan KV)
	defer close(proposeC)
	kvs := NewKVStore(proposeC)
	log.Println("created kv store", kvs)

	go func() {
		for {
			select {
			case data := <-proposeC:
				log.Println("propose kv", data)
				be.Propose(context.TODO(), Entry{ Data: data })
			case entries := <-be.applyThread:
				log.Println("apply kv", entries)
				// TODO write as wrapper?
				for _, entry := range entries {
					data := entry.Data.(KV)
					kvs.Set(data.Key, data.Val)
				}
			}
		}
	}()


	go func() {
		log.Println("set in kv") 
		kvs.ProposeSet("a", "1")
		log.Println("set in kv done") 
		//kvs.Set("b", "2")
		//log.Println("finish set in kv") 
	}()

	//for {}
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)
	errorC := make(chan error)
	defer close(confChangeC)

	// taken directly from https://github.com/etcd-io/etcd/blob/main/contrib/raftexample/httpapi.go
	log.Println("serving http.....")
	serveHTTPKVAPI(kvs, *kvport, confChangeC, errorC)
}
