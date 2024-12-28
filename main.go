package main

import (
	"fmt"
)

func main() {
	sharedLog := NewFakeVirtualLog()
	localStore := NewFakeLocalStore()
	be := NewBaseEngine(sharedLog, localStore)

	fmt.Println(be)
}
