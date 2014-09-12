# jsoncfgo

A Go library for reading configuration settings from JSON files.

## Installation

``` bash
$ go get github.com/go-goodies/go_jsoncfg
```

## Usage

jsoncfgo can handle the following data types from the .json configuration file that it reads:

* bool
* int
* int64
* string
* []string
* []int64

You can also call the __Validate__ function to validate attempts to read __non-existent__ variables.

### jsconfgo Advanced Features

* Read __environment variables__
* Handle __included file__ objects that refer to other config files
* Handle __nested json objects__ within the config file
* __Validation__:
  *  Validate existence of __required__ variables
  *  Validate attempt to read __non-existent__ variable

## Usage

### simple-config.json
``` javascript
{
   "host": "localhost",
   "port": 5432,
   "bignNumber": 999999999999999,
   "active": true,
   "appList": ["myapp", "sharedapp"],
   "numbers": [9, 8, 7, 6]
}
```

### simple_config.go

``` go
package main

import (
	"fmt"
	"time"
	"log"
	"github.com/go-goodies/go_jsoncfg"
)

func main() {

	cfg := jsoncfgo.Load("/Users/lex/dev/go/data/jsoncfgo/simple-config.json")

	host := cfg.String("host")
	fmt.Printf("host: %v\n", host)
	bogusHost := cfg.String("bogusHost", "default_host_name")
	fmt.Printf("host: %v\n\n", bogusHost)

	port := cfg.Int("port")
	fmt.Printf("port: %v\n", port)
	bogusPort := cfg.Int("bogusPort", 9000)
	fmt.Printf("bogusPort: %v\n\n", bogusPort)

	bigNumber := cfg.Int64("bignNumber")
	fmt.Printf("bigNumber: %v\n", bigNumber)
	bogusBigNumber := cfg.Int64("bogusBigNumber", 9000000000000000000)
	fmt.Printf("bogusBigNumber: %v\n\n", bogusBigNumber)

	active := cfg.Bool("active")
	fmt.Printf("active: %v\n", active)
	bogusFalseActive := cfg.Bool("bogusFalseActive", false)
	fmt.Printf("bogusFalseActive: %v\n", bogusFalseActive)
	bogusTrueActive := cfg.Bool("bogusTrueActive", true)
	fmt.Printf("bogusTrueActive: %v\n\n", bogusTrueActive)

	appList := cfg.List("appList")
	fmt.Printf("appList: %v\n", appList)
	bogusAppList := cfg.List("bogusAppList", []string{"app1", "app2", "app3"})
	fmt.Printf("bogusAppList: %v\n\n", bogusAppList)

	numbers := cfg.IntList("numbers")
	fmt.Printf("numbers: %v\n", numbers)
	bogusSettings := cfg.IntList("bogusSettings", []int64{1, 2, 3})
	fmt.Printf("bogusAppList: %v\n\n", bogusSettings)

	if err := cfg.Validate(); err != nil {
		time.Sleep(100 * time.Millisecond)
		defer log.Fatalf("ERROR - Invalid config file...\n%v", err)
		return
	}
}
```

### Output

``` text
host: localhost
host: default_host_name

port: 5432
bogusPort: 9000

bigNumber: 999999999999999
bogusBigNumber: 9000000000000000000

active: true
bogusFalseActive: false
bogusTrueActive: true

appList: [myapp sharedapp]
bogusAppList: [app1 app2 app3]

numbers: [9 8 7 6]
bogusAppList: [1 2 3]

2014/07/25 19:25:03 ERROR - Invalid config file...
Multiple errors: Missing required config key "bogusAppList" (list of strings), Missing required config key "bogusSettings" (list of ints)
exit status 1

Process finished with exit code 1
```

## Notes

See interface documenation at [package jsoncfgo] (http://godoc.org/github.com/go-goodies/go_jsoncfg)

See companion article at [jsoncfgo - A JSON Config File Reader] (http://l3x.github.io/golang-code-examples/2014/07/25/jsoncfgo-config-file-reader-advanced.html)

For a code example of how to use some of the advanced features of jsoncfgo, see [jsoncfgo - Advanced Usage] (http://l3x.github.io/golang-code-examples/2014/07/25/jsoncfgo-config-file-reader-advanced.html)

