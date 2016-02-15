rcon
====
[![Build Status](https://travis-ci.org/madcitygg/rcon.svg)](https://travis-ci.org/madcitygg/rcon)
[![Test Coverage](https://img.shields.io/codecov/c/github/madcitygg/rcon.svg)](https://codecov.io/github/madcitygg/rcon)
[![License](https://img.shields.io/github/license/madcitygg/rcon.svg)](https://github.com/madcitygg/rcon/blob/master/LICENSE.md)

This package implements the `rcon` protocol for communicating with Source engine servers.

Usage
-----
A simple example:
```go
package main

import (
    "github.com/madcitygg/rcon"
)

func main() {
    r, err := rcon.Dial("10.10.10.10:27015")
    if err != nil {
        panic(err)
    }
    defer r.Close()

    err = r.Authenticate("password")
    if err != nil {
        panic(err)
    }

    response, err := r.Execute("status")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Response: %+v\n", response)
}
```
