package kv

import (
	bytes
)

// KV state machine

type KVState struct {
	kvs map[string][]byte
}

func (s *KVState) put() {

}

func (s *KVState) apply(op []byte)[]byte {
	// first 3 bytes tell us what operation it is
	opid := op[:3]
	op = op[3:]
	if bytes.Equal(opid, []byte("PUT")) {
		// do put
	} else if bytes.Equal(opid, []byte("GET")) {
		// do get
	}
}
