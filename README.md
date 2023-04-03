# Fly With Golang

A distributed flight system built using golang and UDP! In this repository contains the code required to run a distributed server that listens and executes based on the request given! The server listens to UDP datagrams and the data representation accepted follows Common Data Representation (CDR) which has been custom implemented (as part of the requirements given). The interface to be given relies on the function selector of each endpoint given in `service.go`.

The app structure is based on the hexagonal architecture. In the `cmd` folder contains the entrypoint, `internal` folder contains the apis and server while `pkg` contains relevant packages that can be used globally. A dockerfile has also been included to support hosting on a server.

First install all dependencies:

```
#Install all dependencies in go.mod
make download

#Else, you can run the comman directly
go mod download
```

To start serving locally:

```
#Run Makefile command
make serve

#Else, you can run the command directly
go run cmd/main.go
```

To serve via docker image:

```
#Build the docker image first
docker build -t <name> .
#e.g
docker build --platform linux/amd64 -t gfs .

#Run the docker image
docker run -dp <port>:8888/udp --user 1001 --name <container-name> <image-name>
#e.g
docker run -dp 8888:8888/udp --user 1001 --name gfs gfs
```
