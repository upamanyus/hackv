package main

import (
	"flag"
	"fmt"
	"github.com/mit-pdos/gokv/grove_ffi"
	"github.com/upamanyus/hackv/kv"
	"os"
)

func main() {
	// var ctrlStr string
	// flag.StringVar(&ctrlStr, "ctrl", "", "address of controller")
	// TODO: should add this

	var port uint64
	flag.Uint64Var(&port, "port", 0, "port number to user for server")
	flag.Parse()

	usage_assert := func(b bool) {
		if !b {
			flag.PrintDefaults()
			fmt.Print("Port number must be included")
			os.Exit(1)
		}
	}

	usage_assert(port != 0)

	me := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", port))
	internalMe := grove_ffi.MakeAddress(fmt.Sprintf("0.0.0.0:%d", 0))
	s := kv.MakeKVServer()
	s.Start(grove_ffi.Address(me), grove_ffi.Address(internalMe))
	select{}
}
