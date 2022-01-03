package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/upamanyus/hackv/urpc/rpc"
	"log"
)

type KVServer struct {
	r  *RSMServer
	k  *KVState
}

func (ks *KVServer) OperationRPC(args []byte, reply *[]byte) {
	log.Printf("Operation started")
	defer log.Printf("Operation completed")
	err, val := ks.r.Op(args)
	*reply = MarshalOpReply(err, val)
}

func MakeKVServer() *KVServer {
	ks := new(KVServer)
	ks.k = MakeKVState()
	ks.r = MakeRSMServer(ks.k.apply)
	return ks
}

func (ks *KVServer) Start(addr grove_ffi.Address, internalAddr grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[RSM_OPERATION] = ks.OperationRPC

	r := rpc.MakeRPCServer(handlers)
	r.Serve(addr)
	ks.r.Start(internalAddr)
}
