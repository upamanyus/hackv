package pb

// XXX: this file is currently unused

import (
	"sync"
)

type PrimaryServer struct {
	cn   uint64
	tlog *TruncatedLog

	replicaClerks []*ReplicaClerk
	nextIndex     []uint64
}

type LogID struct {
	index uint64
	cn    uint64
}

func (s *PrimaryServer) TryAppend(e LogEntry) LogID {
	index := s.tlog.append(e, s.cn)
	for i, ck := range s.replicaClerks {
		localCk := ck
		localRid := uint64(i)
		args := &AppendLogArgs{
			Tlog:s.tlog.suffix(s.nextIndex[i]),
			Cn:s.cn,
		}
		go func () {
			reply := new(AppendLogReply)
			localCk.AppendLog(args, reply)
			s.postAppendLog(localRid, reply)
			// s.postAppendLog(localRid, reply)
		}()
	}
	return LogID{index: index, cn: s.cn}
}

func (s *PrimaryServer) postAppendLog(rid uint64, r *AppendLogReply) {
	if r.Success {
		s.nextIndex[rid] = r.LastIndex + 1
	} else {
		s.nextIndex[rid] = r.LastIndex
	}
}

// Takes over ownership of tlog
func MakePrimaryServer(tlog *TruncatedLog, conf *PBConfiguration) *PrimaryServer {
	s := new(PrimaryServer)
	s.tlog = tlog
	s.cn = conf.cn
	s.replicaClerks = make([]*ReplicaClerk, len(conf.replicas)-1)
	s.nextIndex = make([]uint64, len(conf.replicas)-1)
	for i, host := range conf.replicas[1:] {
		s.replicaClerks[i] = MakeReplicaClerk(host)
		s.nextIndex[i] = tlog.highestIndex()
	}
	return s
}

// XXX: This is a bad abstraction.
// This is really two things in one: a PrimaryLearnerServer that gets accepted
// witnesses, and a (Follower)LearnerServer that just finds out about ever
// increasing commitIndex values.
type LearnerServer struct {
	acceptedIndex []uint64
	commitIndex   uint64
	cond          *sync.Cond
}

func MakeLearnerServer(cond *sync.Cond) *LearnerServer {
	s := new(LearnerServer)
	s.acceptedIndex = nil
	s.commitIndex = 0
	s.cond = cond
	return s
}

func min(l []uint64) uint64 {
	var m uint64 = uint64(18446744073709551615)
	for _, v := range l {
		if v < m {
			m = v
		}
	}
	return m
}

func (s *LearnerServer) postAppendLog(rid uint64, r *AppendLogReply) {
	if r.Success {
		if r.LastIndex > s.acceptedIndex[rid] {
			s.acceptedIndex[rid] = r.LastIndex
			newCommitIndex := min(s.acceptedIndex)
			if newCommitIndex > s.commitIndex {
				s.commitIndex = newCommitIndex
				s.cond.Signal()
			}
		}
	}
}

func (s *LearnerServer) MakePrimaryLearner(conf *PBConfiguration) {
	s.acceptedIndex = make([]uint64, len(conf.replicas))
	for i, _ := range s.acceptedIndex {
		s.acceptedIndex[i] = 0
	}
}
