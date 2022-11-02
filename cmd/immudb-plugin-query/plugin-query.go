package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/go-hclog"
	immudbStorage "github.com/hlexx/jaeger-immudb/pkg/immudb-storage"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"os"
)

var configPath string

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "jaeger",
		Level:      hclog.Warn,
		JSONFormat: true,
	})
	flag.StringVar(&configPath, "config", "cfg/plugin-config.yaml", "The absolute path to the plugin's configuration file")
	flag.Parse()
	logger.Warn("Init immudb config file")
	driver, err := immudbStorage.New(configPath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load config file %v", err.Error()))
		os.Exit(1)
	}
	plugin, err := immudbStorage.NewStoreQuery(driver, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed init storage %v", err.Error()))
		os.Exit(1)
	}
	logger.Warn("Init service")
	grpc.Serve(&shared.PluginServices{
		Store: plugin,
	})
}
