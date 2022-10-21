package cached

import (
	"fmt"
	prolog "git.mills.io/prologic/bitcask"
	"github.com/zenthangplus/goccm"
	"os"
	"sync"
	"time"
)

var (
	MaxPartCashSize  = 1024 * 1024 * 100
	MaxValueCashSize = 1024 * 1024 * 50
	DATABASE         = map[string]*Data{}
	mtx              sync.Mutex
)

type Data struct {
	db *prolog.Bitcask
}
type Item struct {
	Key   []byte
	Value []byte
}

//Connect init data
func Connect(path string) (*Data, error) {
	mtx.Lock()
	defer mtx.Unlock()
	if DATABASE[path] == nil {
		tmpPath := fmt.Sprintf("%s/%s", os.TempDir(), path)
		base, err := prolog.Open(tmpPath, prolog.WithMaxDatafileSize(MaxPartCashSize), prolog.WithMaxValueSize(uint64(MaxValueCashSize)), prolog.WithAutoRecovery(true))
		if err != nil {
			return nil, err
		}
		DATABASE[path] = &Data{
			db: base,
		}
	}
	return DATABASE[path], nil
}

//Add value by key
func (data *Data) Add(key string, value []byte) error {
	return data.db.Put([]byte(key), value)
}

func (data *Data) AddWithTTL(key string, value []byte, ttl time.Duration) error {
	return data.db.PutWithTTL([]byte(key), value, ttl)
}

//Get value by key
func (data *Data) Get(key string) ([]byte, error) {
	mtx.Lock()
	defer mtx.Unlock()
	return data.db.Get([]byte(key))
}

//Exist validate key
func (data *Data) Exist(key string) ([]byte, error) {
	mtx.Lock()
	defer mtx.Unlock()
	get, err := data.db.Get([]byte(key))
	if err != nil && err.Error() == "error: key not found" {
		return get, nil
	} else {
		return get, nil
	}
}

//Remove value by key
func (data *Data) Remove(key string) error {
	return data.db.Delete([]byte(key))
}

//GetAll get all values
func (data *Data) GetAll() (map[string][]byte, error) {
	result := map[string][]byte{}
	err := data.db.Fold(func(key []byte) error {
		get, err := data.db.Get(key)
		if err != nil {
			return err
		}
		result[string(key)] = get
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (data *Data) GetValues() (<-chan Item, error) {
	totalScan := 0
	concurrency := goccm.New(10)
	valuesChannel := make(chan Item)
	go func(inputCache *Data) {
		defer close(valuesChannel)
		_ = data.db.Fold(func(key []byte) error {
			totalScan++
			concurrency.Wait()
			go func(resp chan Item) {
				defer concurrency.Done()
				get, err := data.db.Get(key)
				if err != nil {
					return
				}
				valuesChannel <- Item{
					Key:   key,
					Value: get,
				}
			}(valuesChannel)
			return nil
		})
		fmt.Printf("Cache total scan: %d\n", totalScan)
		concurrency.WaitAllDone()
	}(data)
	return valuesChannel, nil
}
