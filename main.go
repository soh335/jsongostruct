package main

import (
	"flag"
	"fmt"
	"github.com/soh335/jsongostruct/jsongostruct"
	"os"
)

var (
	name = flag.String("name", "XXX", "struct name")
)

func main() {
	flag.Parse()

	if err := jsongostruct.Convert(os.Stdin, os.Stdout, *name); err != nil {
		fmt.Print(err)
	}
}
