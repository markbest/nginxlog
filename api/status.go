package api

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/markbest/nginxlog/conf"
	"github.com/markbest/nginxlog/utils"
	"net/http"
)

//search status data
func GetStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	//log handle
	logFile, logHandle := utils.ElasticLogHandle()
	defer logFile.Close()

	//parse request params
	params := utils.NewHttpPrams(r)
	page := params.GetPage()
	perPage := params.GetPerPage()
	status := params.GetStatus()

	//elastic client
	esClient := utils.NewES(conf.Conf.Elastic.ElasticUrl, logHandle)
	res, _ := esClient.Index(conf.Conf.Elastic.ElasticIndex).
		Type(conf.Conf.Elastic.ElasticType).
		Where("status", status).
		OrderBy("created_at", "desc").
		Take(perPage).
		Page(page).
		Search()

	var logs []utils.LogFormat
	for _, l := range res.Logs {
		logs = append(logs, *l)
	}
	result, err := json.Marshal(logs)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(result))
}
