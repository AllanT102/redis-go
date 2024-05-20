package storage

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "sync"
    "time"
)

// KeyValueStore struct will hold the in-memory data and a mutex for concurrency control
type KeyValueStore struct {
    sync.Mutex
    store map[string]string
    expiration map[string]int64
    filePath string
}

// NewKeyValueStore initializes a new KeyValueStore
func NewKeyValueStore(filePath string) *KeyValueStore {
    kv := &KeyValueStore{
        store:    make(map[string]string),
        expiration: make(map[string]int64),
        filePath: filePath,
    }
    kv.Load() // Load existing data from file if it exists
    return kv
}

// Set a key with a value in the KeyValueStore
func (kv *KeyValueStore) Set(key, value string, timeValid int64) {
    kv.Lock()
    kv.store[key] = value
    expirationTime := timeValid
    if timeValid != -1 {
        expirationTime += time.Now().UnixMilli()
    }
    kv.expiration[key] = expirationTime
    kv.Unlock()
    kv.Save() // Save after each set
}

// Get retrieves a value for a key from the KeyValueStore
func (kv *KeyValueStore) Get(key string) (string, bool) {
    kv.Lock()
    value, ok := kv.store[key]
    if ok {
        expTime, _ := kv.expiration[key]
        if expTime != -1 && time.Now().After(time.UnixMilli(expTime)) {
            kv.Unlock()
            kv.Delete(key)
            kv.Lock()
            value, ok = "", false
        }
    }
    kv.Unlock()
    return value, ok
}

// Delete removes a key from the KeyValueStore
func (kv *KeyValueStore) Delete(key string) {
    kv.Lock()
    delete(kv.store, key)
    delete(kv.expiration, key)
    kv.Unlock()
    kv.Save() // Save after deletion
}

// Save writes the current state of the store to a file
func (kv *KeyValueStore) Save() {
    kv.Lock()
    defer kv.Unlock()
    data, err := json.Marshal(kv.store)
    if err != nil {
        fmt.Println("Error marshaling data:", err)
        return
    }
    err = ioutil.WriteFile(kv.filePath, data, 0644)
    if err != nil {
        fmt.Println("Error writing to file:", err)
    }
}

// Load reads the store's state from a file
func (kv *KeyValueStore) Load() {
    data, err := ioutil.ReadFile(kv.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Println("No existing data file found.")
            return
        }
        fmt.Println("Error reading from file:", err)
        return
    }
    err = json.Unmarshal(data, &kv.store)
    if err != nil {
        fmt.Println("Error unmarshaling data:", err)
    }
}
