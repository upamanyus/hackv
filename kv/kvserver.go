package kv

import (
	"github.com/upamanyus/hackv/pb"
	"sync"
)

type Callback struct {
	cond *sync.Cond
	ret  Error
}

type KVServer struct {
	mu *sync.Mutex
	r  *pb.ReplicaServer

	lastAppliedIndex uint64

	// FIXME: should use async RPC interface
	callbacks map[pb.LogID]Callback
}

type Error = uint64

func (ks *KVServer) applyThread() {
	for {
		le := new(pb.LogEntryCn)
		err := ks.r.GetLogEntry(ks.lastAppliedIndex+1, le)
		if err != pb.ENone {
			continue
		}
		// le.CN
	}
}

func (ks *KVServer) Put(key []byte, val []byte) Error {
	ks.mu.Lock()
	err, l := ks.r.TryAppend(MarshalPutOp(key, val))
	lid := pb.LogID{}
	if err != pb.ENone {
		return err
	}

	c := ks.callbacks[lid]
	ks.mu.Unlock()
}
