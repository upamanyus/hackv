package pb

import (
	"sync"
)

type LogEntryCn struct {
	e  LogEntry
	cn uint64
}

type TruncatedLog struct {
	firstIndex uint64
	log        []LogEntryCn
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func (t *TruncatedLog) highestIndex() uint64 {
	return t.firstIndex + uint64(len(t.log)) - 1
}

// requires: t.firstIndex <= j <= t.highestIndex()
func (t *TruncatedLog) lookupIndex(j uint64) LogEntryCn {
	return t.log[j-t.firstIndex]
}

// requires: t.firstIndex <= j <= t.highestIndex()
// returns the log from index j (inclusive) up to the end of `t`
func (t *TruncatedLog) tailFrom(index uint64) []LogEntryCn {
	return t.log[index-t.firstIndex:]
}

// requires: t.firstIndex <= j <= t.highestIndex()
// returns the log from index firstIndex up to index j (exclusive)
func (t *TruncatedLog) subseqTo(index uint64) []LogEntryCn {
	return t.log[:index-t.firstIndex]
}

func (t *TruncatedLog) truncate(index uint64) {
	t.log = t.tailFrom(index)
	t.firstIndex = index
}

// requires:
// own_TruncatedLog t l
// (exists cn', cn' <= cn && accepted(cn', me, l))
// proposal_lb(cn, l')
// ensures:
// own_TruncatedLog t l || own_TruncatedLog t l' && accepted(cn, me, l')
//
// There's basically two separate proofs for this: one for when cn' == cn, and
// one for when cn' < cn.
func (t *TruncatedLog) tryAppend(a TruncatedLog) bool {
	if t.firstIndex > a.highestIndex() {
		return false // tlog contains entries strictly after `entries`
	}

	if a.firstIndex > t.highestIndex() {
		return false // `entries` contains entries strictly after tlog
	}

	// at this point, t.log and `entries` have some overlap. Let's check if the
	// lowest index that they both contain matches
	indexToCheck := max(t.firstIndex, a.firstIndex)

	if a.lookupIndex(indexToCheck).cn != t.lookupIndex(indexToCheck).cn {
		return false // didn't match
	}

	j := indexToCheck + 1

	for j < a.highestIndex() {
		if j > t.highestIndex() {
			// `entries` has more log entries than t.log
			t.log = append(t.log, t.tailFrom(j)...)
		}
		if t.lookupIndex(j).cn != t.lookupIndex(j).cn {
			// overwrite t.log
			t.log = append(t.subseqTo(j), a.tailFrom(j)...)
		}
	}
	return true
}

type ReplicaServer struct {
	mu *sync.Mutex

	cn   uint64
	tlog TruncatedLog

	commitIndex  uint64
	appliedIndex uint64
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

	reply.Success = s.tlog.tryAppend(args.tlog)

	// XXX: this would make the mutex invariant simpler, but is not necessary
	// for correctness
	// if reply.Success {
	// s.cn = args.Cn
	// }

	reply.LastIndex = s.tlog.highestIndex()
}

func (s *ReplicaServer) GetNextLogEntry() LogEntry {
	s.mu.Lock()
	defer s.mu.Lock()

	for s.appliedIndex > s.commitIndex {
		// s.commit_cond.Wait()
	}

	ret := s.tlog.lookupIndex(s.commitIndex).e
	s.commitIndex += 1
	return ret
}
