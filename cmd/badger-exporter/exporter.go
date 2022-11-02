package main

import (
	"flag"
	"fmt"
	badgerV3 "github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/go-hclog"
	immudbStore "github.com/hlexx/jaeger-immudb/pkg/immudb-storage"
	"github.com/hlexx/jaeger-immudb/pkg/utils"
	"io/ioutil"
	"os"
	"time"
)

var (
	originalPath = "bdata"
	configPath   string
	sleepTime    = time.Second * 10
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "jaeger",
		Level:      hclog.Warn,
		JSONFormat: true,
	})
	flag.StringVar(&configPath, "config", "cfg/plugin-config.yaml", "The absolute path to the plugin's configuration file")
	flag.Parse()
	driver, err := immudbStore.New(configPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load config file %v", err.Error()))
		return
	}
	logger.Warn("Init immudb config file")
	for {
		logger.Warn("init tempDir")
		file, err := ioutil.TempDir(os.TempDir(), "bdata")
		if err != nil {
			logger.Error(fmt.Sprintf("failed create tempfile %v", err.Error()))
			time.Sleep(sleepTime)
			continue
		}
		logger.Warn("copy badger dir")
		err = utils.CopyDir(originalPath, file)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to copy dir %v", err.Error()))
			time.Sleep(sleepTime)
			os.RemoveAll(file)
			continue
		}
		path := fmt.Sprintf("%s/key", file)
		logger.Warn("export data  to immudb")
		opts := badgerV3.DefaultOptions(path)
		opts.SyncWrites = false
		opts.ValueDir = fmt.Sprintf("%s/value", file)
		store, err := badgerV3.Open(opts)
		if err != nil {
			logger.Error(fmt.Sprintf("failed badger open %v", err.Error()))
			time.Sleep(sleepTime)
			os.RemoveAll(file)
			continue
		}
		err = driver.ImportFromBackup(store)
		if err != nil {
			logger.Error(fmt.Sprintf("failed import from backup %v", err.Error()))
			time.Sleep(sleepTime)
			os.RemoveAll(file)
			continue
		}
		os.RemoveAll(file)
	}
}
