# bootstrap_go
Lightweight bootstrap nodes for Golem network. Basic implementation in Go.

# installation

## prerequisites 
- install `go` tools https://golang.org/doc/install  
- create go workspace `mkdir ~/go`

## get package and deps

```
go get github.com/golemfactory/bootstrap_go
cd ~/go/src/github.com/golemfactory/bootstrap_go/
# get dependencies and test dependencies
go get -v -t ./...
```

## run
```
go run main/main.go 
```

## tests

```
go test -v -race ./...
```
where `-race` enables builtin race detector 

## benchmarks

```
go test -bench=. -benchtime=20s ./...
```

# development

## get code
```
git clone git@github.com:golemfactory/bootstrap_go.git
```

[![CircleCI](https://circleci.com/gh/golemfactory/bootstrap_go.svg?style=svg)](https://circleci.com/gh/golemfactory/bootstrap_go)
