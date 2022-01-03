package kv

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/tchajed/marshal"
	"github.com/upamanyus/hackv/urpc/rpc"
)

type RSMClerk struct {
	cl *rpc.RPCClient
}

const (
	RSM_OPERATION = uint64(1)
)

func (ck *RSMClerk) Operation(args []byte) (Error, []byte) {
	raw_reply := new([]byte)
	ck.cl.Call(RSM_OPERATION, args, raw_reply, 100 /* ms */)
	return UnmarshalOpReply(*raw_reply)
}

func MarshalOpReply(err uint64, val []byte) []byte {
	enc := marshal.NewEnc(8 + uint64(len(val)))
	enc.PutInt(err)
	enc.PutBytes(val)
	return enc.Finish()
}

func UnmarshalOpReply(raw_reply []byte) (err uint64, val []byte) {
	dec := marshal.NewDec(raw_reply)
	return dec.GetInt(), raw_reply[8:]
}

func MakeRSMClerk(addr grove_ffi.Address) RSMClerk {
	return RSMClerk{rpc.MakeRPCClient(addr)}
}
