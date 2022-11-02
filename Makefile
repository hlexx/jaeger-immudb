GO=${GOROOT}/bin/go
JAEGER_QUERY=temp-data/jaeger-query
JAEGER_COLLECTOR=temp-data/jaeger-collector

git-update:
	git rm -rf --cached .
	git add .
local-c-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux ${GO} build -ldflags '-extldflags "-static"' -o temp-data/exporter cmd/badger-exporter/exporter.go
local-q-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux ${GO} build -ldflags '-extldflags "-static"' -o temp-data/query cmd/immudb-plugin-query/plugin-query.go
query-run:
	${JAEGER_QUERY} --query.ui-config=ui.json --span-storage.type=grpc-plugin --grpc-storage-plugin.binary=temp-data/query --grpc-storage-plugin.configuration-file=cfg/plugin-config.yaml --metrics-backend=none --log-level=debug
query-run-help:
	${JAEGER_QUERY} env
collector-run-help:
	${JAEGER_COLLECTOR} -h
