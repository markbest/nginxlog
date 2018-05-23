package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"github.com/markbest/nginxlog/api"
	"github.com/markbest/nginxlog/conf"
	"github.com/markbest/nginxlog/process"
	"github.com/markbest/nginxlog/utils"
	"runtime"
	"sync"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// load config
	if err := conf.InitConfig(); err != nil {
		panic(err)
	}

	// monitor parse log process
	go parseLogProcess()

	// auto clear logs
	go utils.ClearLogs()

	// start http server
	startHttpServer()
}

func parseLogProcess() {
	tick := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-tick.C:
			// log handle
			logFile, logHandle := utils.ElasticLogHandle()
			defer logFile.Close()

			// elastic client
			esClient := utils.NewES(conf.Conf.Elastic.ElasticUrl, logHandle)
			esClient.CreateIndex(conf.Conf.Elastic.ElasticIndex)

			wgp := &sync.WaitGroup{}
			targetLogFile := utils.GetLast10MinLogFile()
			count := utils.GetLogsDataCount(targetLogFile)
			if count > 0 {
				wgp.Add(count)
			} else {
				continue
			}

			reader := &process.ReadFromFile{
				FilePath: targetLogFile,
			}

			writer := &process.WriteToES{
				ESClient: esClient,
				ESIndex:  conf.Conf.Elastic.ElasticIndex,
				ESType:   conf.Conf.Elastic.ElasticType,
				Wgp:      wgp,
			}

			logProcess := process.LogProcess{
				ReadChan:  make(chan string),
				WriteChan: make(chan string),
			}

			go logProcess.ReadSource(reader)
			go logProcess.ParseLogData()
			go logProcess.WriteTarget(writer)
			wgp.Wait()
		}
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
