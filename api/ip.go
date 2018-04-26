package api

import (
	"github.com/markbest/nginxlog/conf"
	"github.com/markbest/nginxlog/utils"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

//search ip data
func GetTopIP(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//log handle
	logFile, logHandle := utils.ElasticLogHandle()
	defer logFile.Close()

	//elastic client
	esClient := utils.NewES(conf.Conf.Elastic.ElasticUrl, logHandle)
	res, _ := esClient.Index(conf.Conf.Elastic.ElasticIndex).
		Type(conf.Conf.Elastic.ElasticType).
		Count("*").
		GroupBy("remote_addr").
		Search()

	rs := make(map[string]string)
	for k := range res.Aggs {
		rs[k] = string(*res.Aggs[k])
	}
	fmt.Fprint(w, rs["groupBy"])
}
