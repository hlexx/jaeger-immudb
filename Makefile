GO=${GOROOT}/bin/go

git-update:
	git rm -rf --cached .
	git add .
local-build:
	${GO} build -ldflags '-extldflags "-static"' -o temp-data/exporter cmd/badger-exporter/exporter.go
