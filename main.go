package main

import (
	"beast/global"
	"beast/protocol"
	"beast/server"
	"beast/util"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Config:", global.Config)
	//pprof
	go func() {
		profServeMux := http.NewServeMux()
		profServeMux.HandleFunc("/debug/pprof/", pprof.Index)
		profServeMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		profServeMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		profServeMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		err := http.ListenAndServe(":9998", profServeMux)
		if err != nil {
			panic(err)
		}
	}()

	//server.InitServer(2, 1024, 3, 60, "127.0.0.1:9999", new(protocol.HttpParser))
	//server.InitServer(2, 1024, 3, 3600, "127.0.0.1:9999", new(protocol.TextParser))
	server.InitServer(2, 1024, 3, 3600, "127.0.0.1:9999", new(protocol.TextParser))
	util.InitSignal()
}
