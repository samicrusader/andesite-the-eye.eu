#!/usr/bin/env bash

init() {
    go install github.com/rakyll/statik
    ~/go/bin/statik -src="./www/" -f
}
build_template() {
    export CGO_ENABLED=1
    export GOOS=$1
    export GOARCH=$2
    export GOARM=7
    ext=$3
    date=$(date +'%Y.%m.%d')
    version=${CIRCLE_BUILD_NUM-$date}
    tag=the-eye.eu-v$version-$(git log --format=%h -1)
    mkdir -p ./bin/andesite-$tag-$GOOS-$GOARCH
    echo $tag-$GOOS-$GOARCH
    go build -ldflags="-s -w -X main.Version=$tag" -o ./bin/andesite-$tag-$GOOS-$GOARCH/andesite$ext
    go build -ldflags="-s -w -X main.Version=$tag" -o ./bin/andesite-$tag-$GOOS-$GOARCH/hashr$ext ./cmd/hashr
    tar -C ./bin -cf ./bin/andesite-$tag-$GOOS-$GOARCH.tar andesite-$tag-$GOOS-$GOARCH
}

init
build_template linux amd64
