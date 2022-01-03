package kv

import (
	"github.com/upamanyus/hackv/pb"
	"sync"
)

type Callback struct {
	cond     *sync.Cond
	complete bool
	reply    []byte
}

type RSMServer struct {
	mu *sync.Mutex
	r  *pb.ReplicaServer

	lastAppliedIndex uint64

	// FIXME: should use async RPC interface
	callbacks map[pb.LogID]*Callback

	apply func([]byte)[]byte
}

type Error = uint64

func (ks *RSMServer) applyThread() {
	for {
		le := new(pb.LogEntryCn)
		err := ks.r.GetLogEntry(ks.lastAppliedIndex+1, le)
		if err != pb.ENone {
			continue
		}
		lid := pb.LogID{Index:ks.lastAppliedIndex+1, Cn:le.Cn}

		reply := ks.apply(le.E)
		ks.mu.Lock()

		// FIXME: want to wake up failed ops too

		if c, ok := ks.callbacks[lid]; ok {
			c.reply = reply
			c.cond.Signal()
		}
		ks.mu.Unlock()
		// le.CN
	}
}

func (ks *RSMServer) Op(op []byte) (Error, []byte) {
	ks.mu.Lock()
	err, lid := ks.r.TryAppend(op)
	if err != pb.ENone {
		return err, nil
	}

	oldCb, ok := ks.callbacks[lid]
	if ok {
		oldCb.cond.Signal()
	}
	cb := &Callback{cond: sync.NewCond(ks.mu), complete: false, reply: nil}
	ks.callbacks[lid] = cb
	for !cb.complete {
		cb.cond.Wait()
	}

	ks.mu.Unlock()
	return pb.ENone, cb.reply
}

func MakeRSMServer() {
	// FIXME: impl
}

func (s *RSMServer) Start(h uint64) {
	// FIXME: impl
}
