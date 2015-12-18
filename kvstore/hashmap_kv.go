package kvstore

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
)

//HashKVStore implements the KVStore interface using a hashmap
type HashKVStore struct {
	faulty      bool
	Store       map[string][]byte
	backingFile string
}

func createBackingFileIfNeeded(filename string) error {
	_, err := os.Stat(filename)
	if err != nil {
		//If there's a problem with the path, we'll try creating the file, otherwise we give up
		switch err.(type) {
		default:
			return err
		case *os.PathError:
			break
		}

		//Attempt to create the file
		f, err := os.Create(filename)
		if err != nil {
			return err
		}

		defer f.Close()
	}

	return nil
}

//NewHashKVStore creates a new HashKVStore
func NewHashKVStore(backingFile string) (*HashKVStore, error) {
	if backingFile != "" {
		log.Info("Using backing store with filename ", backingFile)
	}

	kvStore := &HashKVStore{
		faulty:      false,
		Store:       make(map[string][]byte),
		backingFile: backingFile,
	}

	if backingFile == "" {
		return kvStore, nil
	}

	err := createBackingFileIfNeeded(backingFile)
	if err != nil {
		return nil, err
	}

	if err := kvStore.LoadFromFile(); err != nil {
		return nil, err
	}

	return kvStore, nil

}

//InjectFaults forces Gets and Puts to return errors
func (hkvs *HashKVStore) InjectFaults() {
	hkvs.faulty = true
}

//ClearFaults resets store to non-fault state
func (hkvs *HashKVStore) ClearFaults() {
	hkvs.faulty = false
}

//Put stores a value under the given key
func (hkvs *HashKVStore) Put(key string, value []byte) error {
	if hkvs.faulty {
		return errors.New("Faulty store does not put ur key/val pair, ok?")
	}
	hkvs.Store[key] = value
	return nil
}

//Get returns the value (if any) under the given key
func (hkvs *HashKVStore) Get(key string) ([]byte, error) {
	if hkvs.faulty {
		return nil, errors.New("Faulty store does not get ur key, ok?")
	}
	return hkvs.Store[key], nil
}

//List returns all the values (if any) stored under the given key
func (hkvs *HashKVStore) List(key string) ([]*KVPair, error) {

	if hkvs.faulty {
		return nil, errors.New("You can haz list? Nope.")
	}

	var kvpairs []*KVPair
	for k, v := range hkvs.Store {
		if strings.HasPrefix(k, key) {
			kvpairs = append(kvpairs, &KVPair{k, v})
		}
	}
	return kvpairs, nil
}

//DumpToFile writes the content of the kv store to the backing file configured for the store.
func (hkvs *HashKVStore) DumpToFile() error {
	if hkvs.backingFile == "" {
		return fmt.Errorf("No path to backing file specified")
	}

	f, err := os.Create(hkvs.backingFile)
	if err != nil {
		return err
	}

	defer f.Close()

	for key, value := range hkvs.Store {
		line := fmt.Sprintf("%s#%s\n", key, value)
		_, err := f.Write([]byte(line))
		if err != nil {
			return err
		}
	}

	return nil

}

//LoadFromFile loads the flushed KVStore representation from file into memory
func (hkvs *HashKVStore) LoadFromFile() error {
	if hkvs.backingFile == "" {
		return fmt.Errorf("No path to backing file specified")
	}

	f, err := os.Open(hkvs.backingFile)
	if err != nil {
		return err
	}
	defer f.Close()

	loadedMap := make(map[string][]byte)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "#")
		if len(parts) != 2 {
			log.Info("Line did not split into two parts - skipping: ", line)
			continue
		}
		loadedMap[parts[0]] = []byte(parts[1])
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	hkvs.Store = loadedMap

	return nil
}

//Flush writes the KVStore in memory representation to disk.
func (hkvs *HashKVStore) Flush() error {
	if hkvs.backingFile == "" {
		log.Info("Flush called on HashKNStore with no backing file - ignoring.")
		return nil
	}

	log.Info("flush called on kvstore")
	return hkvs.DumpToFile()
}
