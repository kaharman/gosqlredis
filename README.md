# gosqlredis

Go library that stores data in Redis with SQL-like schema. The goal of this library is we can store data in Redis with table form.

[![Go Reference](https://pkg.go.dev/badge/github.com/kaharman/gosqlredis.svg)](https://pkg.go.dev/github.com/kaharman/gosqlredis)

## What you need

1. Golang struct with Redis tag
2. Data that you want to store

## Supported data type

1. Standard Golang data type
2. ```time``` data type
2. ```database/sql``` data type

## How to use

   ```go get -u github.com/kaharman/gosqlredis```

## License

gosqlredis is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).
