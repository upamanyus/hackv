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
		log.Fatalf("Unexpected error %d", err)
	}
}

func (ck *KVClerk) Get(key []byte) []byte {
	err, val := ck.rsm.Operation(MarshalGetOp(key))
	if err != pb.ENone {
		log.Fatalf("Unexpected error %d", err)
	}
	return val
}
