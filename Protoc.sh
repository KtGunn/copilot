#!/bin/bash

MODULE=`head -1 go.mod | awk '{print $2}'`
protoc -I./proto --go_opt=module=${MODULE} --go_out=. --go-grpc_opt=module=${MODULE} --go-grpc_out=. proto/*.proto
