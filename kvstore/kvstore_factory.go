package kvstore

import (
	"fmt"
	"net/url"
)

//NewKVStore instantiates a KV store implementation based on the url scheme associated with the given url
func NewKVStore(envURL string) (KVStore, error) {
	u, err := url.Parse(envURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	default:
		return nil, fmt.Errorf("Unrecognized scheme for KVStore URL: %v", u.Scheme)
	case "consul":
		return makeConsulKVStore(u)
	case "file":
		return makeHashMapKVStore(u)
	}
}

func makeConsulKVStore(u *url.URL) (KVStore, error) {
	ckvs, err := NewConsulKVStore(u)
	if err != nil {
		return nil, err
	}
	return KVStore(ckvs), nil
}

func makeHashMapKVStore(u *url.URL) (KVStore, error) {
	hkvs, err := NewHashKVStore(u.Path)
	if err != nil {
		return nil, err
	}
	return KVStore(hkvs), nil
}
