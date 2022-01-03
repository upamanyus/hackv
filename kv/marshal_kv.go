package kv

import (
	"github.com/tchajed/marshal"
)

const (
	PUT_OP = uint64(1)
	GET_OP = uint64(2)
)

func MarshalPutOp(key []byte, val []byte) []byte {
	// XXX: don't actually need 8 bytes for the lengths, can set a max length with fewer bytes
	enc := marshal.NewEnc(3 + 8 + uint64(len(key)) + 8 + uint64(len(val)))
	enc.PutBytes([]byte("PUT"))
	enc.PutInt(uint64(len(key)))
	enc.PutBytes(key)
	enc.PutInt(uint64(len(val)))
	enc.PutBytes(val)
	return enc.Finish()
}

func UnmarshalPutOp(op []byte) (key []byte, val []byte) {
	dec := marshal.NewDec(op)
	key = dec.GetBytes(dec.GetInt())
	val = dec.GetBytes(dec.GetInt())
	return
}

func MarshalGetOp(key []byte) []byte {
	enc := marshal.NewEnc(3 + uint64(len(key)))
	enc.PutBytes([]byte("GET"))
	enc.PutBytes(key)
	return enc.Finish()
}

func UnmarshalGetOp(op []byte) []byte {
	return op
}
