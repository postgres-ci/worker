#!/bin/bash

rm -rf tmp && mkdir tmp

go build -ldflags='-s' -o tmp/worker ../worker.go

docker build -t postgres-ci-worker .