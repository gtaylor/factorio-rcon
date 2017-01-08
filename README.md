factorio-rcon
=============
[![GoDoc](https://godoc.org/github.com/gtaylor/factorio-rcon?status.svg)](https://godoc.org/github.com/gtaylor/factorio-rcon)
[![License](https://img.shields.io/github/license/gtaylor/factorio-rcon.svg)](https://github.com/gtaylor/factorio-rcon/blob/master/LICENSE.md)

This package is a fork of [madcitygg/rcon](https://github.com/madcitygg/rcon) with a few tweaks to work with Factorio. Namely, Factorio's rejection of sending `SERVERDATA_RESPONSE_VALUE` packets to check for multi-packet responses (details [here](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#Multiple-packet_Responses)).

Usage
-----
A simple example:
```go
package main

import (
    "fmt"
    "github.com/gtaylor/factorio-rcon"
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

License
-------

Like the upstream [madcitygg/rcon](https://github.com/madcitygg/rcon), this package is licensed under the MIT License.
