# rcon

This package implements the `rcon` protocol for communicating with Source engine servers.

## Usage

A simple example:
```go
package main

import (
    "github.com/madcitygg"
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
