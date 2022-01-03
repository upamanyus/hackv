package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
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

	apply func([]byte) []byte
}

type Error = uint64

func (ks *RSMServer) applyThread() {
	for {
		le := new(pb.LogEntryCn)
		err := ks.r.GetLogEntry(ks.lastAppliedIndex+1, le)
		if err != pb.ENone {
			continue
		}
		lid := pb.LogID{Index: ks.lastAppliedIndex + 1, Cn: le.Cn}

		reply := ks.apply(le.E)
		ks.lastAppliedIndex += 1
		ks.mu.Lock()

		// FIXME: want to wake up failed ops too

		if cb, ok := ks.callbacks[lid]; ok {
			cb.complete = true
			cb.reply = reply
			cb.cond.Signal()
		}
		ks.mu.Unlock()
	}
}

func (ks *RSMServer) Op(op []byte) (Error, []byte) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
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

	return pb.ENone, cb.reply
}

func MakeRSMServer(apply func([]byte) []byte) *RSMServer {
	s := new(RSMServer)
	s.mu = new(sync.Mutex)
	s.r = pb.MakeReplicaServer()

	s.lastAppliedIndex = 0

	s.callbacks = make(map[pb.LogID]*Callback)
	s.apply = apply
	return s
}

func (s *RSMServer) Start(h grove_ffi.Address) {
	go s.applyThread()
	s.r.Start(h)
}
