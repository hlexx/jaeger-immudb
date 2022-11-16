package main

import (
	"context"
	"flag"
	"fmt"
	badgerV3 "github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/go-hclog"
	immudbStore "github.com/hlexx/jaeger-immudb/pkg/immudb-storage"
	"github.com/hlexx/jaeger-immudb/pkg/utils"
	"io/ioutil"
	"math"
	"os"
	"time"
)

var (
	originalPath = "bdata"
	sleepTime    = time.Second * 10
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "cfg/plugin-config.yaml", "The absolute path to the plugin's configuration file")
	flag.Parse()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "jaeger",
		Level:      hclog.Warn,
		JSONFormat: true,
	})
	driver, err := immudbStore.New(configPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load config file %v", err.Error()))
		return
	}
	logger.Warn("Init immudb config file")
	for {
		select {
		case <-time.After(time.Second * 30):
			logger.Warn("service work")
		default:
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
				err := os.RemoveAll(file)
				if err != nil {
					logger.Error(fmt.Sprintf("path %v remove all %v", file, err.Error()))
				}
				continue
			}
			path := fmt.Sprintf("%s/key", file)
			logger.Warn("export data  to immudb")
			opts := badgerV3.DefaultOptions(path)
			opts.SyncWrites = false
			opts.ValueDir = fmt.Sprintf("%s/value", file)
			opts.NumVersionsToKeep = math.MaxInt32
			store, err := badgerV3.Open(opts)
			if err != nil {
				logger.Error(fmt.Sprintf("failed badger open %v", err.Error()))
				time.Sleep(sleepTime)
				err := os.RemoveAll(file)
				if err != nil {
					logger.Error(fmt.Sprintf("path %v remove all %v", file, err.Error()))
				}
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
			db := &immudbStore.BadgerDB{
				Db:      store,
				Context: ctx,
			}
			go func() {
				err = driver.ImportFromBackup(db)
				if err != nil {
					logger.Error(fmt.Sprintf("failed import from backup %v", err.Error()))
					time.Sleep(sleepTime)
					err := os.RemoveAll(file)
					if err != nil {
						logger.Error(fmt.Sprintf("path %v remove all %v", file, err.Error()))
					}
					return
				}
				err = os.RemoveAll(file)
				if err != nil {
					logger.Error(fmt.Sprintf("path %v remove all %v", file, err.Error()))
				}
				logger.Warn("remove dir")
			}()
			<-ctx.Done()
			cancel()
			errC := ctx.Err()
			logger.Warn(fmt.Sprintf("backup context done Err:[%v]", errC))
			err = os.RemoveAll(file)
			if err != nil {
				logger.Error(fmt.Sprintf("path %v remove all %v", file, err.Error()))
			}
			time.Sleep(sleepTime)
		}
	}
}
