// Package gridscale is a Go library to consume the Gridscale API (http://www.gridscale.de).
// The source is hosted on GitHub at https://github.com/parce-iot/gridscale
//
// Usage:
//    package main
//
//    import (
//        "github.com/parce-iot/gridscale"
//
//        "fmt"
//        "os"
//    )
//
//    func main() {
//        userID := os.Getenv("GRIDSCALE_USERID")
//        apiToken := os.Getenv("GRIDSCALE_APITOKEN")
//        endpoint := "https://api.gridscale.io"
//
//        c, err := gridscale.NewClient(userID, apiToken, endpoint)
//        if err != nil {
//            fmt.Printf("ERROR: %s", err)
//            return
//        }
//
//        c.GetPrices()
//    }
package gridscale
