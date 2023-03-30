# Fly With Golang

A distributed flight system built using golang and UDP! In this repository contains the code required to run a distributed server that listens and executes based on datagram input given!

The app structure is based on the hexagonal architecture. In the `cmd` folder contains the entrypoint, `internal` folder contains the apis and server while `pkg` contains relevant packages that can be used globally.

Data representations are largely based on the COBRA data representation. Interfaces are based on the function selectors that can be found in `service.go`

First install all dependencies:

```
#Install all dependencies in go.mod
make download

#Else, you can run the comman directly
go mod download
```

To start serving:

```
#Run Makefile command
make serve

#Else, you can run the command directly
go run cmd/main.go
```
