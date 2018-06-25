package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/markbest/nginxlog/api"
	"github.com/markbest/nginxlog/conf"
	"github.com/markbest/nginxlog/process"
	"github.com/markbest/nginxlog/utils"

	"github.com/julienschmidt/httprouter"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// load config
	if err := conf.InitConfig(); err != nil {
		panic(err)
	}

	// monitor parse log process
	go parseLogProcess()

	// debug pprof
	debugpprof()

	// start http server
	startHttpServer()
}

func parseLogProcess() {
	tick := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-tick.C:
			// auto clear logs
			utils.ClearLogs()

			// log handle
			logFile, logHandle := utils.ElasticLogHandle()

			// elastic client
			esClient := utils.NewES(conf.Conf.Elastic.ElasticUrl, logHandle)

			endFlagChan := make(chan int)
			targetLogFile := utils.GetLast10MinLogFile()
			reader := &process.ReadFromFile{
				FilePath: targetLogFile,
			}

			writer := &process.WriteToES{
				ESClient:    esClient,
				ESIndex:     conf.Conf.Elastic.ElasticIndex,
				ESType:      conf.Conf.Elastic.ElasticType,
				EndFlagChan: endFlagChan,
			}

			logProcess := process.LogProcess{
				ReadChan:  make(chan string, 100),
				WriteChan: make(chan string, 100),
			}

			go logProcess.ReadSource(reader)
			go logProcess.ParseLogData()
			go logProcess.WriteTarget(writer)
			<-endFlagChan
			logFile.Close()
		}
	}
}

// start debug pprof server
func debugpprof() {
	if conf.Conf.App.Debug {
		pprofServer := &http.Server{Addr: conf.Conf.App.Pprof}
		go pprofServer.ListenAndServe()
	}
}

// start http server
func startHttpServer() {
	router := httprouter.New()
	router.GET("/api/analysis/status", api.GetStatus)
	router.GET("/api/analysis/method", api.GetMethod)
	router.GET("/api/analysis/topIp", api.GetTopIP)
	log.Fatal(http.ListenAndServe(conf.Conf.App.Port, router))
}
