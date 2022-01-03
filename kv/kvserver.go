package kv

import (
	"sync"
)

type KVServer struct {
	mu *sync.Mutex
	r  *RSMServer
}

func (ks *KVServer) PutRPC(args []byte, reply []byte) {

}

func (ks *KVServer) GetRPC(key []byte, reply []byte) {

}

// Should we even bother making them separate RPCs?
// What about a single type of RPC called "Operation"?
// Don't want to marshal/unmarshal back and forth.
