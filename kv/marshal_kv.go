package kv

const (
	PUT_OP = uint64(1)
	GET_OP = uint64(2)
)

func MarshalPutOp(key []byte, val []byte) []byte {
	enc := marshal.NewEnc(3 + uint64(len(key)) + uint64(len(val)))
	enc.PutBytes([]byte("PUT"))
	enc.PutBytes([]byte("GET"))
	return enc.Finish()
}
