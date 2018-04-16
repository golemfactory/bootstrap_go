# bootstrap_go
Lightweight bootstrap nodes for Golem network. Basic implementation in Go.

# installation

## prerequisites 
- install `go` tools https://golang.org/doc/install  
- create go workspace `mkdir ~/go` https://github.com/golang/go/wiki/SettingGOPATH

## get code and deps

This will get the sources into `GOPATH` which is by default `~/go`.

```
go get github.com/golemfactory/bootstrap_go
cd ~/go/src/github.com/golemfactory/bootstrap_go/
# get dependencies and test dependencies
go get -t ./...
```

## run
```
go run main/main.go 
```

## tests

```
go test ./...
```

## benchmarks

```
go test -bench=. -benchtime=20s ./...
```


[![CircleCI](https://circleci.com/gh/golemfactory/bootstrap_go.svg?style=svg)](https://circleci.com/gh/golemfactory/bootstrap_go)
