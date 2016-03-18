Golang gridscale SDK
====================

[![GoDoc](https://godoc.org/github.com/Shopify/sarama?status.png)](https://godoc.org/github.com/parce-iot/gridscale)

Introduction
------------
This is a Go (Golang) library to consume the Gridscale API
(http://www.gridscale.de).

[Malte Janduda](https://github.com/MalteJ) from [Parce](http://www.parce.de) has
initiated this project.

This code is not officially supported by Gridscale.

All coders are invited to use this library and send pull requests!


Usage
-----

    package main
    
    import (
        "github.com/parce-iot/gridscale"

        "fmt"
        "os"
    )

    func main() {
        userID := os.Getenv("GRIDSCALE_USERID")
        apiToken := os.Getenv("GRIDSCALE_APITOKEN")
        endpoint := "https://api.gridscale.io"

        c, err := gridscale.NewClient(userID, apiToken, endpoint)
        if err != nil {
            fmt.Printf("ERROR: %s", err)
            return
        }
        
        c.GetPrices()
    }

License
-------
This library is under the Apache License v2.0. For more information have a look
into the [LICENSE](LICENSE) file. When contributing to this project (e.g. by
sending pull requests) you accept your work to be under the same license. To
show your consent please git sign your commits (`git commit -s -m 'foo'`).

We only accept signed commits!


Development
-----------
### General
Please sign your git commits to show your consent to put your work under the
Apache License v2.

Please go lint your code before committing and sending pull requests.

Please keep your pull requests small! Send multiple pull requests for bigger
changes! Use the GitHub issues to talk to us beforehand.


### Testing
Clone this repository into your GOPATH:

    go get github.com/parce-iot/gridscale
    cd $GOPATH/src/github.com/parce-iot/gridscale

At first you have to provide your gridscale API credentials as environment
variables:

    export GRIDSCALE_USERID=c702de16-a89f-4edb-b05a-09c7593ac65a
    export GRIDSCALE_APITOKEN=5df171ddd552b7c47fc67c83100c06e1268021b2a2e5827d535ddef7333fe64b

then change into the project directory `gridscale` and execute

    make test

### Dependencies
This library uses and ships with the [shopspring/decimal](https://github.com/shopspring/decimal)
library (MIT License) to handle fixed-point decimal numbers.

TODO
----
* Read object events
* Create Storage Snapshots
* Locations
* Templates
* ISO images
* Deleted objects
