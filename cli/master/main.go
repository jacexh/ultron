package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/qastub/ultron"
	"go.uber.org/zap"
)

var (
	masterListen string
	webListen    string
)

func dump(w http.ResponseWriter, o interface{}) {
	body, err := ultron.J.Marshal(o)
	if err != nil {
		dump(w, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func main() {
	flag.StringVar(&masterListen, "master", ":9500", "MasterRunner listening port")
	flag.StringVar(&webListen, "web", ":9600", "the web api listening port")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.NotFound(w, req)
			return
		}

		baser := new(ultron.BaseRunner)
		err := ultron.J.NewDecoder(req.Body).Decode(baser)
		if err != nil {
			dump(w, map[string]string{"error": err.Error()})
			return
		}

		fmt.Println("baserunner: ", baser)
		fmt.Println("config: ", baser.Config)

		if ultron.MasterRunner.GetStatus() == ultron.StatusBusy {
			dump(w, map[string]string{"error": "MasterRunner is running"})
			return
		}

		ultron.MasterRunner.WithConfig(baser.Config)
		ultron.MasterRunner.WithDeadLine(baser.Deadline)
		fmt.Println("ultron.MasterRunner.WithDeadLine(baser.Deadline)", ultron.MasterRunner.Deadline)
		ultron.ServerStart <- struct{}{}

		dump(w, map[string]string{"msg": "ok"})
		return
	})

	mux.HandleFunc("/stop", func(w http.ResponseWriter, req *http.Request) {
		if ultron.MasterRunner.GetStatus() == ultron.StatusBusy {
			ultron.ServerStop <- struct{}{}
			dump(w, map[string]string{"msg": "stopped"})
			return
		}

		dump(w, map[string]string{"error": "not running"})
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		dump(w, map[string]string{"hello": "world"})
	})

	go func() {
		lis, err := net.Listen("tcp", masterListen)
		if err != nil {
			panic(err)
		}
		ultron.MasterRunner.Listener = lis
		fmt.Println(ultron.MasterRunner)
		ultron.MasterRunner.Start()
	}()

	ultron.Logger.Info("web api listen on port " + webListen)
	ultron.Logger.Panic("panic", zap.Error(http.ListenAndServe(webListen, mux)))
}
