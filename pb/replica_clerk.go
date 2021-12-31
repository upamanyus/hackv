package pb

import (
	"github.com/mit-pdos/gokv/urpc/rpc"
	"github.com/mit-pdos/gokv/grove_ffi"
)

type ReplicaClerk struct {
	cl *rpc.RPCClient
	conf []rpc.HostName
}

// ReplicaClerk implements ReplicaServerRPCs

func MakeReplicaClerk(host grove_ffi.Address) *ReplicaClerk {
	return nil // FIXME: impl
}

func (ck *ReplicaClerk) AppendLog(args *AppendLogArgs, reply *AppendLogReply) {
	args_raw := EncodeAppendLogArgs(args)
	raw_reply := new([]byte)
	err := ck.cl.Call(REPLICA_APPEND, args_raw, raw_reply, 100 /* ms */ )
	if err == 0 {
		DecodeAppendLogReply(*raw_reply, reply)
	} else {
		reply = nil
	}
}

func (ck *ReplicaClerk) UpdateCommitIndex(args *UpdateCommitIndexArgs) {
	args_raw := EncodeUpdateCommitIndexArgs(args)
	raw_reply := new([]byte)
	_ = ck.cl.Call(REPLICA_UPDATECOMMIT, args_raw, raw_reply, 100 /* ms */ )
}

func (ck *ReplicaClerk) BecomePrimary(args *BecomePrimaryArgs) {
	args_raw := EncodeBecomePrimaryArgs(args)
	raw_reply := new([]byte)
	_ = ck.cl.Call(REPLICA_APPEND, args_raw, raw_reply, 100 /* ms */ )
}
