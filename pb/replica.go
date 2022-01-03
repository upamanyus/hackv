package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/upamanyus/hackv/urpc/rpc"
	"sync"
)

type PBConfiguration struct {
	cn       uint64
	replicas []grove_ffi.Address
}

type ReplicaServer struct {
	mu *sync.Mutex

	cn         uint64
	tlog       *TruncatedLog
	acceptedCn uint64 // XXX: needed for UpdateCommitIndex()
	// Strictly speaking, a bool is enough, but this makes the code simpler.
	// With a bool, we'd have to invalidate in any place cn might increase.
	// With this, we're being conservative and it'll automatically become
	// invalid when cn increases.

	commitIndex uint64
	applyCond   *sync.Cond

	isPrimary bool
	// Primary state
	replicaClerks []*ReplicaClerk
	nextIndex     []uint64
	acceptedIndex []uint64
}

func (s *ReplicaServer) TruncateLog(Index uint64) {
	s.mu.Lock()
	s.tlog.truncate(Index)
	s.mu.Unlock()
}

func (s *ReplicaServer) AppendLog(args *AppendLogArgs, reply *AppendLogReply) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reply.LastIndex = s.tlog.highestIndex()
	reply.Success = false

	if args.Cn < s.cn {
		reply.Cn = s.cn
		return
	}

	s.cn = args.Cn // might enter new conf *without* accepting anything

	reply.Success = s.tlog.tryAppend(args.Tlog)
	if reply.Success {
		s.acceptedCn = args.Cn
	}
	reply.LastIndex = s.tlog.highestIndex()
	reply.Cn = s.cn
}

const (
	ENone       = uint64(0)
	ETruncated  = uint64(1)
	ENotPrimary = uint64(2)
)

func (s *ReplicaServer) GetLogEntry(index uint64, e *LogEntryCn) Error {
	s.mu.Lock()
	defer s.mu.Lock()

	for index > s.commitIndex && index >= s.tlog.firstIndex {
		s.applyCond.Wait()
	}

	if index < s.tlog.firstIndex {
		return ETruncated
	}
	*e = s.tlog.lookupIndex(s.commitIndex)
	return ENone
}

func (s *ReplicaServer) TryAppend(e LogEntry) (Error, LogID) {
	s.mu.Lock()
	defer s.mu.Lock()
	if !s.isPrimary {
		return ENotPrimary, LogID{}
	}

	index := s.tlog.append(e, s.cn)
	for i, ck := range s.replicaClerks {
		localCk := ck
		localRid := uint64(i)
		args := &AppendLogArgs{
			Tlog: s.tlog.suffix(s.nextIndex[i]),
			Cn:   s.cn,
		}
		go func() {
			reply := new(AppendLogReply)
			localCk.AppendLog(args, reply)
			s.postAppendLog(localRid, reply)
		}()
	}

	return ENone, LogID{index: index, cn: s.cn}
}

func (s *ReplicaServer) postAppendLog(rid uint64, reply *AppendLogReply) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !reply.Success {
		s.nextIndex[rid] = reply.LastIndex
	} else {
		s.nextIndex[rid] = reply.LastIndex + 1
		if reply.LastIndex > s.acceptedIndex[rid] {
			s.acceptedIndex[rid] = reply.LastIndex
			newCommitIndex := min(s.acceptedIndex)
			if newCommitIndex > s.commitIndex {
				s.commitIndex = newCommitIndex
				s.applyCond.Signal()
			}
		}
	}
}

func (s *ReplicaServer) UpdateCommitIndex(cn uint64, index uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cn <= s.acceptedCn { // If we've accepeted past cn, then we have the same entries
		if index > s.commitIndex {
			s.commitIndex = index
		}
	}
}

func (s *ReplicaServer) BecomePrimary(args *BecomePrimaryArgs) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cn >= args.Conf.cn {
		return
	}

	// init primary state
	s.replicaClerks = make([]*ReplicaClerk, len(args.Conf.replicas)-1)
	s.nextIndex = make([]uint64, len(args.Conf.replicas)-1)
	s.acceptedIndex = make([]uint64, len(args.Conf.replicas))

	s.acceptedIndex[0] = s.tlog.highestIndex()
	for i, host := range args.Conf.replicas[1:] {
		s.replicaClerks[i] = MakeReplicaClerk(host)
		s.nextIndex[i] = s.tlog.highestIndex()
		s.acceptedIndex[i] = 0
	}
}

func MakeReplicaServer() *ReplicaServer {
	s := new(ReplicaServer)
	s.mu = new(sync.Mutex)
	s.cn = 0
	s.tlog = MakeTruncatedLog()
	s.applyCond = sync.NewCond(s.mu)

	return s
}

func (s *ReplicaServer) Start(host grove_ffi.Address) {
	handlers := make(map[uint64]func([]byte, *[]byte))
	handlers[REPLICA_APPEND] = func(raw_args []byte, raw_reply *[]byte) {
		args := new(AppendLogArgs)
		reply := new(AppendLogReply)
		DecodeAppendLogArgs(raw_args, args)
		s.AppendLog(args, reply)
		*raw_reply = EncodeAppendLogReply(reply)
	}

	handlers[REPLICA_UPDATECOMMIT] = func(raw_args []byte, raw_reply *[]byte) {
		args := new(UpdateCommitIndexArgs)
		DecodeUpdateCommitIndexArgs(raw_args, args)
		s.UpdateCommitIndex(args.Cn, args.CommitIndex)
		*raw_reply = nil
	}

	handlers[REPLICA_BECOMEPRIMARY] = func(raw_args []byte, raw_reply *[]byte) {
		args := new(BecomePrimaryArgs)
		DecodeBecomePrimaryArgs(raw_args, args)
		s.BecomePrimary(args)
		*raw_reply = nil
	}
	rpc.MakeRPCServer(handlers).Serve(host)
}
