package pb

import (
	"github.com/mit-pdos/gokv/grove_ffi"
	"sync"
)

type PBConfiguration struct {
	cn       uint64
	replicas []grove_ffi.Address
}

type ReplicaServer struct {
	mu *sync.Mutex

	cn         uint64
	tlog       TruncatedLog
	acceptedCn uint64 // XXX: needed for UpdateCommitIndex()
	// Strictly speaking, a bool is enough, but this makes the code simpler.
	// With a bool, we'd have to invalidate in any place cn might increase.
	// With this, we're being conservative and it'll automatically become
	// invalid when cn increases.

	commitIndex uint64
	applyCond  *sync.Cond

	p *PrimaryServer
	l *LearnerServer
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

	reply.Success = s.tlog.tryAppend(args.tlog)
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

func (s *ReplicaServer) GetLogEntry(index uint64, e *LogEntry) Error {
	s.mu.Lock()
	defer s.mu.Lock()

	for index > s.commitIndex && index >= s.tlog.firstIndex {
		s.applyCond.Wait()
	}

	if index < s.tlog.firstIndex {
		return ETruncated
	}
	*e = s.tlog.lookupIndex(s.commitIndex).e
	return ENone
}

func (s *ReplicaServer) TryAppend(e LogEntry) (Error, LogID) {
	s.mu.Lock()
	defer s.mu.Lock()
	if s.p == nil {
		return ENotPrimary, LogID{}
	} else {
		return ENone, s.p.TryAppend(e)
	}
}

func (s *ReplicaServer) UpdateCommitIndex(cn uint64, index uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cn <= s.acceptedCn {
		if index > s.l.commitIndex {
			s.l.commitIndex = index
		}
	}
}

func (s *ReplicaServer) BecomePrimary(conf *PBConfiguration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cn >= conf.cn {
		return
	}
	s.p = MakePrimaryServer(s.tlog.clone(), conf)
	s.l.MakePrimaryLearner(conf)
}

func MakeReplicaServer() *ReplicaServer {
	return nil
}

func (s *ReplicaServer) Start(h uint64) {

}
