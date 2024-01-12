#!/bin/sh

set -xue

DATE="$(date +'%Y-%m-%d')"
mkdir -p target
rm target/* || true

env GOOS=linux GOARCH=amd64 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_linux_amd64_"${DATE}" ./cmd
env GOOS=linux GOARCH=386 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_linux_amd64_"${DATE}" ./cmd
env GOOS=linux GOARCH=arm64 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_linux_arm64_"${DATE}" ./cmd
env GOOS=linux GOARCH=arm go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_linux_arm_"${DATE}" ./cmd

env GOOS=darwin GOARCH=amd64 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_darwin_amd64_"${DATE}" ./cmd
env GOOS=darwin GOARCH=arm64 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_darwin_arm64_"${DATE}" ./cmd

env GOOS=windows GOARCH=amd64 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_windows_amd64_"${DATE}.exe" ./cmd
env GOOS=windows GOARCH=386 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_windows_386_"${DATE}.exe" ./cmd
env GOOS=windows GOARCH=arm64 go build -ldflags '-extldflags "-static"' -tags netgo,osusergo -o target/tikmeh_windows_arm64_"${DATE}.exe" ./cmd
