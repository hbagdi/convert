# convert

[![Docs](https://pkg.go.dev/badge/github.com/hbagdi/convert)](https://pkg.go.dev/github.com/hbagdi/convert)
![Test](https://github.com/hbagdi/convert/workflows/CI%20Test/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/hbagdi/convert)](https://goreportcard.com/report/github.com/hbagdi/convert)

Convert from one type to another.

## Features

- Copy from one struct type to another
- Copy from one field to another field with a different name
- Use custom transformation functions to copy fields of different types

## Usage

```go
package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/hbagdi/convert"
)

type From struct {
	// ID will be copied into To.ID as is.
	ID string
	// CreatedAt will be copied into To.CreationTime.
	// Registered function will be invoked to transform time.Time into int64.
	CreatedAt time.Time `convert:"CreationTime"`
	// Count will copied into Total field.
	// Registered function will be invoked to transform string into int.
	Count string `convert:"Total"`
}

type To struct {
	ID           string
	CreationTime int64
	Total        int
}

func init() {
	convert.Register(reflect.TypeOf(time.Time{}), reflect.TypeOf(int64(0)),
		func(from interface{}) (interface{}, error) {
			t, ok := from.(time.Time)
			if !ok {
				return 0, nil
			}
			return t.Unix(), nil
		},
	)
	convert.Register(
		reflect.TypeOf(""),
		reflect.TypeOf(int(0)),
		func(from interface{}) (interface{}, error) {
			s, ok := from.(string)
			if !ok {
				return "", nil
			}
			return strconv.Atoi(s)
		},
	)
}

func main() {
	f := From{
		ID:        "some-id",
		CreatedAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
		Count:     "42",
	}

	var t To
	err := convert.Convert(f, &t)
	if err != nil {
		log.Fatalln("failed to convert:", err)
	}
	fmt.Printf("From struct: %+v\n", f)
 	// From struct: {ID:some-id CreatedAt:2030-01-01 00:00:00 +0000 UTC Count:42}
	fmt.Printf("To struct: %+v\n", t)
 	// To struct: {ID:some-id CreationTime:1893456000 Total:42}
}
```
## Acknowledgement

This package borrows code from [github.com/jinzhu/copier](https://github.com/jinzhu/copier).

## License

convert is licensed with Apache License Version 2.0.
Please read the [LICENSE](LICENSE) file for more details.
