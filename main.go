package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if os.Args[3] == "en" {
		compressFile()
	} else {
		t1 := time.Now()
		deCompressFile()

		fmt.Println("Time decode: ", time.Now().Sub(t1))
	}

}
