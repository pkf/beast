package main

import (
	//"beast/global"
	"net/http"
	"net/http/pprof"
)

func main() {
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

	InitServer(1024, 3, 60, "127.0.0.1:9999")
}
