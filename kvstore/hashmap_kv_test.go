package kvstore

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

type TestThing struct {
	Foo string
	Far int
	Faz string
}

func TestPutAndGet(t *testing.T) {
	kvs, _ := NewHashKVStore("")

	tt := &TestThing{"foo1", 1, "bar1"}
	bytes, err := json.Marshal(tt)
	assert.Nil(t, err)

	err = kvs.Put("foo1", bytes)
	assert.Nil(t, err)

	bytes, err = kvs.Get("foo1")
	assert.Nil(t, err)
	assert.NotNil(t, bytes)

	rehydrated := new(TestThing)
	err = json.Unmarshal(bytes, &rehydrated)
	assert.Equal(t, "foo1", rehydrated.Foo)
}

func TestList(t *testing.T) {
	hkvs, err := NewHashKVStore("")
	assert.Nil(t, err)
	kvs := KVStore(hkvs)

	tt1 := &TestThing{"foo1", 1, "bar1"}
	bytes, err := json.Marshal(tt1)
	assert.Nil(t, err)

	err = kvs.Put("foos/foo1", bytes)
	assert.Nil(t, err)

	//Add it a second time under another key so the returned
	//slice has a larger capabity than the matched elements
	//in the list call
	err = kvs.Put("another/foo1", bytes)
	assert.Nil(t, err)

	tt2 := &TestThing{"foo2", 2, "bar2"}
	bytes, err = json.Marshal(tt2)
	assert.Nil(t, err)

	err = kvs.Put("foos/foo2", bytes)
	assert.Nil(t, err)

	kvpairs, err := kvs.List("foos")
	assert.Nil(t, err)
	assert.NotNil(t, kvpairs)
	assert.Equal(t, 2, len(kvpairs))
}

func TestGetHashMapKVSFromFactory(t *testing.T) {
	//Make a test file
	t.Log("Create temp file")
	f, err := ioutil.TempFile("./", "tst")
	assert.Nil(t, err)

	//Schedule cleanup
	t.Log("Schedule clean up of temp file")
	defer os.Remove(f.Name())

	//Create a hashmap keystore
	t.Log("Create hashmap kv store")
	wd, _ := os.Getwd()

	kvs, err := NewKVStore(fmt.Sprintf("file:///%s/%s", wd, f.Name()))
	assert.Nil(t, err)
	assert.NotNil(t, kvs)
}

func TestGetConsulKVSFromFactory(t *testing.T) {

	//Create a hashmap keystore
	t.Log("Create hashmap kv store")

	kvs, err := NewKVStore("consul://localhost:1")
	assert.Nil(t, err)
	assert.NotNil(t, kvs)
}

func TestKVSFromFactoryMalformedURL(t *testing.T) {

	//Create a hashmap keystore
	t.Log("Create hashmap kv store")

	kvs, err := NewKVStore(":abc")
	assert.NotNil(t, err)
	assert.Nil(t, kvs)
}

func TestKVSFromFactoryUnrecognizedScheme(t *testing.T) {

	//Create a hashmap keystore
	t.Log("Create hashmap kv store")

	kvs, err := NewKVStore("foo://xxx")
	assert.NotNil(t, err)
	assert.Nil(t, kvs)
}

func TestDumpAndLoad(t *testing.T) {
	//Make a test file
	t.Log("Create temp file")
	f, err := ioutil.TempFile("./", "tst")
	assert.Nil(t, err)

	//Schedule cleanup
	t.Log("Schedule clean up of temp file")
	defer os.Remove(f.Name())

	//Create a hashmap keystore
	t.Log("Create hashmap kv store")
	kvs, err := NewHashKVStore(f.Name())
	assert.Nil(t, err)

	//Add some stuff
	t.Log("Populate store")
	tt1 := &TestThing{"foo1", 1, "bar1"}
	bytes, err := json.Marshal(tt1)
	assert.Nil(t, err)

	err = kvs.Put("foos/foo1", bytes)
	assert.Nil(t, err)

	tt2 := &TestThing{"foo2", 2, "bar2"}
	bytes, err = json.Marshal(tt2)
	assert.Nil(t, err)

	err = kvs.Put("foos/foo2", bytes)
	assert.Nil(t, err)

	//Dump it
	t.Log("Dump file")
	err = kvs.DumpToFile()
	assert.Nil(t, err)

	//Load it
	t.Log("Load written file")
	err = kvs.LoadFromFile()
	assert.Nil(t, err)

	t.Log("Ensure loaded map contents equals dumped map")
	assert.Equal(t, 2, len(kvs.Store))

	v, err := kvs.Get("foos/foo1")
	assert.Nil(t, err)
	assert.NotNil(t, v)

	var testThing TestThing
	err = json.Unmarshal(v, &testThing)
	assert.Nil(t, err)
	assert.Equal(t, tt1, &testThing)

	v, err = kvs.Get("foos/foo2")
	assert.Nil(t, err)
	assert.NotNil(t, v)

	err = json.Unmarshal(v, &testThing)
	assert.Nil(t, err)
	assert.Equal(t, tt2, &testThing)

}

func TestLoadOnCreate(t *testing.T) {
	//Make a test file
	t.Log("Create temp file")
	f, err := ioutil.TempFile("./", "tst")
	assert.Nil(t, err)

	//Schedule cleanup
	t.Log("Schedule clean up of temp file")
	defer os.Remove(f.Name())

	t.Log("Write a line to the file")
	f.WriteString(`foos/foo1#{"Foo":"foo1","Far":1,"Faz":"bar1"}`)
	f.WriteString("\n")
	f.Close()

	t.Log("Create hashmap kv store")
	kvs, err := NewHashKVStore(f.Name())
	assert.Nil(t, err)

	t.Log("Verify file data loaded")
	v, err := kvs.Get("foos/foo1")
	assert.Nil(t, err)
	assert.NotNil(t, v)

	var testThing TestThing
	err = json.Unmarshal(v, &testThing)
	assert.Nil(t, err)
	assert.Equal(t, "foo1", testThing.Foo)
	assert.Equal(t, 1, testThing.Far)
	assert.Equal(t, "bar1", testThing.Faz)

	kvs.InjectFaults()
	_, err = kvs.Get("foos/foo1")
	assert.NotNil(t, err)
	err = kvs.Put("foos/foo7", []byte("foo7"))
	assert.NotNil(t, err)
	_, err = kvs.List("foos/foo1")
	assert.NotNil(t, err)
	kvs.ClearFaults()

	kvs.Flush()

}

func TestLoadNoFile(t *testing.T) {
	const filename = "./nofile"
	t.Log("remove nofile if it exists to start clean")
	os.Remove(filename)

	t.Log("Create hashmap kv store with a non-existant backing file")
	_, err := NewHashKVStore(filename)
	assert.Nil(t, err)

	t.Log("Verify backing file has been created")
	_, err = os.Stat(filename)

	t.Log("clean up the file")
	os.Remove(filename)
}

func TestBenignFlushWithNoBackingStore(t *testing.T) {
	t.Log("Create hashmap kv store with a no backing file")
	kvs, err := NewHashKVStore("")
	assert.Nil(t, err)

	t.Log("flush it")
	err = kvs.Flush()
	assert.Nil(t, err)
}
