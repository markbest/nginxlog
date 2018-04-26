package main

import (
	"flag"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"github.com/markbest/nginxlog/api"
	"github.com/markbest/nginxlog/conf"
	"github.com/markbest/nginxlog/utils"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	//load config
	if err := conf.InitConfig(); err != nil {
		log.Panic(err)
	}

	//log handle
	logFile, logHandle := utils.ElasticLogHandle()
	defer logFile.Close()

	//elastic client
	esClient := utils.NewES(conf.Conf.Elastic.ElasticUrl, logHandle)
	esClient.CreateIndex(conf.Conf.Elastic.ElasticIndex)

	//monitor parse log file
	go parseSingleFile(esClient, false)

	//auto clear logs
	go utils.ClearLogs()

	//start http server
	startHttpServer()
}

//monitor parse single log file
func parseSingleFile(client *utils.ES, all bool) {
	tick := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-tick.C:
			t := time.Now()
			if t.Minute()%5 != 0 {
				log.Println("current time:", time.Unix(t.Unix(), 0).Format("2006-01-02 15:04:05"))
				continue
			}
			var logParse utils.LogStreamData
			err := logParse.OpenStream(conf.Conf.Log.TargetPath + conf.Conf.Log.TargetFilePrefix)
			if err != nil {
				log.Println(err)
				continue
			}
			logParse.ParseStream(client, conf.Conf.Elastic.ElasticIndex, conf.Conf.Elastic.ElasticType, all)
			logParse.CloseStream()
		}
	}
}

//start http server
func startHttpServer() {
	//init route
	router := httprouter.New()
	router.GET("/api/analysis/status", api.GetStatus)
	router.GET("/api/analysis/method", api.GetMethod)
	router.GET("/api/analysis/topIp", api.GetTopIP)
	log.Fatal(http.ListenAndServe(conf.Conf.App.Port, router))
}
