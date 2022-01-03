package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/upamanyus/hackv/pb"
	"log"
)

type KVClerk struct {
	rsm RSMClerk
}

func MakeKVClerk(addr grove_ffi.Address) KVClerk {
	return KVClerk{rsm: MakeRSMClerk(addr)}
}

func (ck *KVClerk) Put(key []byte, val []byte) {
	err, _ := ck.rsm.Operation(MarshalPutOp(key, val))
	if err != pb.ENone {
		log.Fatalf("Unexpected clerk error %d", err)
	}
}

func (ck *KVClerk) Get(key []byte) []byte {
	err, val := ck.rsm.Operation(MarshalGetOp(key))
	if err != pb.ENone {
		log.Fatalf("Unexpected clerk error %d", err)
	}
	return val
}

// TODO:
// A real KVClerk will be a clerk to the configuration server (for now, the
// pbcontroller) and will make low-level RPCClients to servers as needed for
// trying operations.
//
// Also need to gracefully give up after a timeout. Possibly, listen for
// notifications from pbctonroller telling us about the new primary.
