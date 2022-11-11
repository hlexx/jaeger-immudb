package immudb_storage

import (
	"bytes"
	"github.com/dgraph-io/badger/v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var SeperatorA = []byte{':'}

func Make(keys ...[]byte) []byte {
	return bytes.Join(keys, SeperatorA)
}

func Test_BackupRestore(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmpdir)
	}()
	s1Path := filepath.Join(tmpdir, "test1")
	s2Path := filepath.Join(tmpdir, "test2")
	opts := badger.DefaultOptions(s1Path)
	opts.Dir = s1Path
	opts.ValueDir = s1Path
	db1, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	prefix := []byte("testsnapshot")
	key1 := Make(prefix, []byte("key1"))
	key2 := Make(prefix, []byte("key2"))
	rawValue := []byte("NotLongValue")
	err = db1.Update(func(tx *badger.Txn) error {
		if err := tx.Set(key1, rawValue); err != nil {
			return err
		}
		return tx.Set(key2, rawValue)
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := byte(0); i < 255; i++ {
		err = db1.Update(func(tx *badger.Txn) error {
			if err := tx.Set(append(key1, i), rawValue); err != nil {
				return err
			}
			return tx.Set(append(key2, i), rawValue)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	var backup bytes.Buffer
	_, err = db1.Backup(&backup, 0)
	if err != nil {
		t.Fatal(err)
	}
	opts = badger.DefaultOptions(s2Path)
	opts.Dir = s2Path
	opts.ValueDir = s2Path
	db2, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	err = db2.Load(&backup, 1)
	if err != nil {
		t.Fatal(err)
	}
}
