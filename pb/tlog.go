package pb

type LogEntryCn struct {
	E  LogEntry
	Cn uint64
}

type TruncatedLog struct {
	firstIndex uint64
	log        []LogEntryCn
}

type Error = uint64

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
	if index <= t.firstIndex {
		return
	}
	if index >= t.highestIndex() {
		index = t.highestIndex()
	}
	t.log = t.tailFrom(index)
	t.firstIndex = index
}

func (t *TruncatedLog) append(e LogEntry, cn uint64) uint64 {
	t.log = append(t.log, LogEntryCn{E:e, Cn:cn})
	return t.firstIndex + uint64(len(t.log)) - 1
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
func (t *TruncatedLog) tryAppend(a *TruncatedLog) bool {
	if t.firstIndex > a.highestIndex() {
		return false // tlog contains entries strictly after `entries`
	}

	if a.firstIndex > t.highestIndex() {
		return false // `entries` contains entries strictly after tlog
	}

	// at this point, t.log and `entries` have some overlap. Let's check if the
	// lowest index that they both contain matches
	indexToCheck := max(t.firstIndex, a.firstIndex)

	if a.lookupIndex(indexToCheck).Cn != t.lookupIndex(indexToCheck).Cn {
		return false // didn't match
	}

	j := indexToCheck + 1

	for j < a.highestIndex() {
		if j > t.highestIndex() {
			// `entries` has more log entries than t.log
			t.log = append(t.log, t.tailFrom(j)...)
		}
		if t.lookupIndex(j).Cn != t.lookupIndex(j).Cn {
			// overwrite t.log
			t.log = append(t.subseqTo(j), a.tailFrom(j)...)
		}
	}
	return true
}

func (t *TruncatedLog) clone() *TruncatedLog {
	newlog := make([]LogEntryCn, len(t.log))
	copy(newlog, t.log)
	return &TruncatedLog{firstIndex:t.firstIndex, log:newlog}
}

func (t *TruncatedLog) suffix(index uint64) *TruncatedLog {
	t2 := new(TruncatedLog)
	t2.firstIndex = index
	t2.log = make([]LogEntryCn, len(t.log))
	copy(t2.log, t.log)
	return t2
}

func MakeTruncatedLog() *TruncatedLog {
	t := new(TruncatedLog)
	t.firstIndex = 0
	t.log = make([]LogEntryCn, 1)
	t.log[0] = LogEntryCn{E:nil, Cn:0}
	return t
}
