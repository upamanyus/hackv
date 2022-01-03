package kv

import (
	"sync"
)

type KVServer struct {
	mu *sync.Mutex
	r  *RSMServer
	k *KVState
}

func (ks *KVServer) OperationRPC(args []byte, reply *[]byte) {
	err, val := ks.r.Op(args)
	*reply = MarshalOpReply(err, val)
}
