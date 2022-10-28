GO=${GOROOT}/bin/go

git-update:
	git rm -rf --cached .
	git add .
local-c-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux ${GO} build -ldflags '-extldflags "-static"' -o temp-data/exporter cmd/badger-exporter/exporter.go
local-q-build:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux ${GO} build -ldflags '-extldflags "-static"' -o temp-data/query cmd/immudb-plugin-query/plugin-query.go
