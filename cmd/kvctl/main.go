package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/upamanyus/hackv/kv"
	"log"
	"os"
)

func main() {
	var ctrlStr string
	flag.StringVar(&ctrlStr, "ctrl", "", "address of controller")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Println("Must provide command in form:")
			fmt.Println(" get KEY")
			fmt.Println(" put KEY VALUE")
			fmt.Println(" add HOST")
			os.Exit(1)
		}
	}

	usage_assert(ctrlStr != "")

	ctrl := grove_ffi.MakeAddress(ctrlStr)
	ck := kv.MakeKVClerk(grove_ffi.Address(ctrl))

	a := flag.Args()
	usage_assert(len(a) > 0)
	if a[0] == "get" {
		usage_assert(len(a) == 2)
		k := []byte(a[1])
		v := ck.Get(k)
		fmt.Printf("GET %d ↦ %v\n", k, v)
	} else if a[0] == "put" {
		usage_assert(len(a) == 3)
		k := []byte(a[1])
		v := []byte(a[2])
		ck.Put(k, v)
		fmt.Printf("PUT %d ↦ %v\n", k, v)
	} else if a[0] == "add" {
		log.Fatalf("Unsupported")
	}
}
