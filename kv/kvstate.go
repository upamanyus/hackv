package kv

import (
	"bytes"
)

// KV state machine

type KVState struct {
	kvs map[string][]byte
}

func (s *KVState) put(key []byte, val []byte) {
	s.kvs[string(key)] = val
}

func (s *KVState) get(key []byte) (val []byte) {
	return s.kvs[string(key)]
}

func (s *KVState) apply(op []byte)[]byte {
	// first 3 bytes tell us what operation it is
	opid := op[:3]
	op = op[3:]
	if bytes.Equal(opid, []byte("PUT")) {
		key, val := UnmarshalPutOp(op)
		s.put(key, val)
	} else if bytes.Equal(opid, []byte("GET")) {
		key := UnmarshalGetOp(op)
		val := s.get(key)
		return val
	} else {
		panic("Unknown op")
	}
	return nil
}
