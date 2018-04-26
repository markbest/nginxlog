package utils

import (
	"fmt"
	"log"
	"github.com/markbest/nginxlog/conf"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//get elastic log handle
func ElasticLogHandle() (logFile *os.File, logHandle *log.Logger) {
	fileName := conf.Conf.Elastic.ElasticLogPath + "/log-" + time.Now().Format("2006-01-02") + ".log"
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		log.Fatalln("open file error !")
	}
	logHandle = log.New(logFile, "[elastic]", log.LstdFlags)
	return logFile, logHandle
}

//get log dir all file
func GetFileList(path string, pattern string) []string {
	files := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		if pattern != "" {
			if strings.Contains(path, pattern) {
				files = append(files, path)
			}
		} else {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	return files
}

//clear es record logs
func ClearLogs() {
	tick := time.NewTicker(time.Hour * 24)
	for {
		select {
		case <-tick.C:
			maxLogFiles := conf.Conf.Elastic.ElasticLogMaxFiles
			logFiles := GetFileList(conf.Conf.Elastic.ElasticLogPath, "")
			if len(logFiles) > 0 {
				for _, v := range logFiles {
					f, _ := os.Stat(v)
					if f.IsDir() {
						continue
					}

					modTime := f.ModTime().Unix()
					curTime := time.Now().Unix()
					if curTime - modTime > int64(maxLogFiles * 86400) {
						os.Remove(v)
					}
				}
			}
		}
	}
}
