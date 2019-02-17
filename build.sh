#!/bin/bash

project="$GOPATH/src/github.com/task_mail/"
 
cd "$project/Server/Room"
go build 
cd "$project/Server"
go install 
cd "$project/Client"
go install 