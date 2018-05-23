package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/markbest/nginxlog/api"
	"github.com/markbest/nginxlog/conf"
	"github.com/markbest/nginxlog/utils"
	"log"
	"net/http"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// load config
	if err := conf.InitConfig(); err != nil {
		log.Panic(err)
	}

	// monitor parse log file
	go parseLogProcess()

	// auto clear logs
	go utils.ClearLogs()

	// start http server
	startHttpServer()
}

// monitor parse single log file
func parseLogProcess() {
	tick := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-tick.C:
			t := time.Now()
			if t.Minute()%5 != 0 {
				log.Println("current time:", time.Unix(t.Unix(), 0).Format("2006-01-02 15:04:05"))
				continue
			}

			// log handle
			logFile, logHandle := utils.ElasticLogHandle()

			// elastic client
			esClient := utils.NewES(conf.Conf.Elastic.ElasticUrl, logHandle)
			esClient.CreateIndex(conf.Conf.Elastic.ElasticIndex)

			var logParse utils.LogStreamData
			err := logParse.OpenStream(conf.Conf.Log.TargetPath + conf.Conf.Log.TargetFilePrefix)
			if err != nil {
				log.Println(err)
				continue
			}
			logParse.ParseStream(esClient, conf.Conf.Elastic.ElasticIndex, conf.Conf.Elastic.ElasticType)
			logParse.CloseStream()
			logFile.Close()
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
