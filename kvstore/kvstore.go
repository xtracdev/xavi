package kvstore

//KVPair represents a key value pair as used by the KVStore interface
type KVPair struct {
	Key   string
	Value []byte
}

//KVStore is implemented by all key value store providers
type KVStore interface {
	Put(string, []byte) error
	Get(string) ([]byte, error)
	List(string) ([]*KVPair, error)
	Flush() error
}
