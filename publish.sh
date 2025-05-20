#!/bin/sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o doc_admin main.go && scp doc_admin root@10.6.0.1:/www/server/docs/
go build -o doc_admin && mv doc_admin ~/mylab/_scripts/

