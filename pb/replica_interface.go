package pb

type LogEntry = []byte

const (
	REPLICA_APPEND        = uint64(1)
	REPLICA_UPDATECOMMIT  = uint64(2)
	REPLICA_BECOMEPRIMARY = uint64(3)
)

type AppendLogArgs struct {
	tlog TruncatedLog
	Cn   uint64
}

type AppendLogReply struct {
	Success   bool
	LastIndex uint64
	Cn        uint64 // In case conf has increased; this will tell primary to talk to server and update conf
}

type UpdateCommitIndexArgs struct {
	Cn          uint64
	CommitIndex uint64
}

type BecomePrimaryArgs struct {
	Cn       uint64
	Replicas []string // all of the replicas, in which we are the first
}

type ReplicaServerRPCs interface {
	// requires: ...
	// ensures: reply.Success == true ==> replica accepted entries in Cn
	AppendLog(args *AppendLogArgs, reply *AppendLogReply)

	// requires: exists log, |log| == CommitIndex && proposal_lb(cn,log) && committed_by(cn, log)
	// ensures: true; only here for progress
	UpdateCommitIndex(args *UpdateCommitIndexArgs)

	// requires: (config_ptsto cn args.Replicas) && (me == Replicas[0])
	// ensures: true; only here for progress
	BecomePrimary(args *BecomePrimaryArgs)
}
