# jaeger-immudb

## Общая информация о проекте
[![License](http://img.shields.io/badge/Licence-MIT-blue.svg)](LICENSE)
![GitHub contributors](https://img.shields.io/github/contributors/hlexx/jaeger-immudb)
[![GoDoc](https://godoc.org/github.com//hlexx/jaeger-immudb?status.svg)](https://godoc.org/github.com/hlexx/jaeger-immudb)
[![Go Report Card](https://goreportcard.com/badge/github.com/hlexx/jaeger-immudb)](https://goreportcard.com/report/github.com/hlexx/jaeger-immudb)

## CI/CD
[![Go](https://github.com/hlexx/jaeger-immudb/actions/workflows/ci.yaml/badge.svg)](https://github.com/hlexx/jaeger-immudb/actions/workflows/ci.yaml)
![Platform](https://img.shields.io/badge/platform-Linux%20-blue)
![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/hlexx/jaeger-immudb)
[![GitHub release](https://img.shields.io/github/v/release/hlexx/jaeger-immudb)](https://github.com/hlexx/jaeger-immudb/releases/latest)

[<img alt="Jaeger" align="left" width="200" height="200" src="https://www.jaegertracing.io/img/jaeger-vector.svg">](https://github.com/jaegertracing/jaeger)
[<img alt="integration" width="200" height="200" src="img/plus.png" width="150" height="150" align="">]()
[<img alt="immudb" src="img/mascot.png" width="200"/>](https://github.com/codenotary/immudb)

---

## Основные возможности:

* Реализована интеграция Jaeger plugin-query через gRPC API c Immudb KV storage. Immudb — это база данных, написанная на Go, позволяет добавлять записи, но не изменять.
* Jaeger collector - собирает трейсы и кладет из во внутреннее хранилище Badger. Далее следует экспорт.
* Экспорт в Immudb происходит параллельно работе Jaeger Collector, через бекап хранилища.
* Plugin-Query забирает данные из Immudb, сюда же входит и веб-интерфейс Jaeger UI.
* Для запуска Immudb нужно открыть порт 3322.

## 🚀 Запуск сервиса


1. Установить пакет
``` bash 
go get github.com/hlexx/jaeger-immudb
   ```

2. Запуск immudb 
``` bash 
docker run -it --rm --name immudb -p 3322:3322 codenotary/immudb:latest
   ```
3.Запуск Jaeger Collector
``` bash 
docker pull ghcr.io/hlexx/jaeger-immudb/collector:latest
   ```
4. Запуск плагина Query
``` bash 
docker pull ghcr.io/hlexx/jaeger-immudb/query:latest
   ```

## Пример добавления трейса в Jaeger
``` go 
package main

import (
   "github.com/google/uuid"
   log "github.com/hlexx/jaeger-immudb/pkg/trace"
)

func main() {
   id := uuid.NewString()
   log.Trace(id, "service", map[string]string{})
}
   ```
