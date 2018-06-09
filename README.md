# Silo

Silo is a bare bones as-simple-as-possible key-data store, where 'data' is simply a []byte of configurable length.

* [Why](#why)
* [Before You Start](#before-you-start)
* [Building and Requirements](#building-and-requirements)
* [ToDo](#todo)

## Why

I needed a simple way to allow a set of workers to read & write file(s) that they can all then fetch for a home project.

## Before You Start

It's important to note that silo isn't intended to act as a per-user store, but rather something like a universal
storage pot used by services.

For example all 'roles' with the 'read' permission can read any key, regardless of who wrote it. Silo doesn't attach
permissions or owners to data blocks. If you want to ensure that only an intended user(s) can read something you
should encrypt the data and forward the cypher text to silo for storage.

## Building and Requirements

The dependencies are straight forward

```go
    github.com/gtank/cryptopasta
    gopkg.in/gcfg.v1
```

Build the server:

```
cd cmd/silo/
go build -o silo *.go
```

## ToDo

* At the moment the data is read into memory when it is received, and then to disk. Really it should be streamed to disk ...
* ^ The same applies in reverse
* More tests ..
* The metrics endpoint (/) could return actual metrics, other than 'ok' as a health check
* At some point there will need to be a layer that routes data to where it is actually saved to allow large
  data blocks to be saved across various disks / hosts. It'll get pretty involved but, and it's too advanced for my
  current use case .. but maybe in future.
