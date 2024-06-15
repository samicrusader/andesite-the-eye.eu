#!/usr/bin/env bash

set -e
set -x

#go test
go build ./cmd/hashr

# refresh sqlite3 db
#touch ~/.config/andesite/files.db
#rm ~/.config/andesite/files.db

# this is about the same config as the main site uses
./hashr --scan-concurrency 8 --verbose --fsdb-verbose --disable-hash sha256 --disable-hash sha512 --disable-hash sha3 --disable-hash blake2b --public $1
