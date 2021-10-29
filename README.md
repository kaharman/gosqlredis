# gosqlredis

[![Go Reference](https://pkg.go.dev/badge/github.com/kaharman/gosqlredis.svg)](https://pkg.go.dev/github.com/kaharman/gosqlredis)
[![Build](https://github.com/kaharman/gosqlredis/actions/workflows/build.yml/badge.svg)](https://github.com/kaharman/gosqlredis/actions/workflows/build.yml)

Go library that stores data in Redis with SQL-like schema. The goal of this library is we can store data in Redis with table form.

## What is ```gosqlredis``` user for
- Cache data effectively without data duplication in Redis
- Store data table in memory database

## How to use

```shell
go get -u github.com/kaharman/gosqlredis
```

## What you need

1. Golang struct with Redis tag
2. Data that you want to store

## Supported data type

1. Standard Golang data type
2. ```time``` data type
2. ```database/sql``` data type

## License

```gosqlredis``` is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
