package process

import (
	"sync"

	"github.com/markbest/nginxlog/utils"
)

type WriteToES struct {
	ESClient *utils.ES
	ESIndex  string
	ESType   string
	Wgp      *sync.WaitGroup
}

// write to es
func (w *WriteToES) Write(writeChan chan string) {
	for data := range writeChan {
		w.ESClient.EsClient.Index().
			Index(w.ESIndex).
			Type(w.ESType).
			BodyJson(data).
			Do()
		w.Wgp.Done()
	}
}
