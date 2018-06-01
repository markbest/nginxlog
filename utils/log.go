package utils

import (
	"bufio"
	"fmt"
	"github.com/markbest/nginxlog/conf"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// nginx log data fields
type LogFormat struct {
	RemoteAddr        string    `json:"remote_addr"`
	RemoteUser        string    `json:"remote_user"`
	TimeLocal         time.Time `json:"time"`
	RequestType       string    `json:"request_type"`
	RequestUrl        string    `json:"request_url"`
	HttpVersion       string    `json:"http_version"`
	Status            int       `json:"status"`
	BodyBytesSent     int       `json:"body_bytes_sent"`
	HttpReferer       string    `json:"http_referer"`
	HttpUserAgent     string    `json:"http_user_agent"`
	HttpXForwardedFor string    `json:"http_x_forwarded_for"`
	CreatedAt         int64     `json:"created_at"`
}

// get elastic log handle
func ElasticLogHandle() (logFile *os.File, logHandle *log.Logger) {
	fileName := conf.Conf.Elastic.ElasticLogPath + "/log-" + time.Now().Format("2006-01-02") + ".log"
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		log.Fatalln("open file error !")
	}
	logHandle = log.New(logFile, "[elastic]", log.LstdFlags)
	return logFile, logHandle
}

// get log dir all file
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

// clear es record logs
func ClearLogs() {
	maxLogFiles := conf.Conf.Elastic.ElasticLogMaxFiles
	t := time.Now().Unix() - int64(maxLogFiles*86400)
	lastF := conf.Conf.Elastic.ElasticLogPath + "/log-" + time.Unix(t, 0).Format("2006-01-02")
	_, err := os.Stat(lastF)
	if !os.IsNotExist(err) {
		os.Remove(lastF)
	}
}

// get log file count lines
func GetLogsDataCount(fileName string) (count int) {
	file, err := os.Open(fileName)
	if err != nil {
		return count
	}
	defer file.Close()

	buff := bufio.NewReader(file)
	for {
		logLine, err := buff.ReadString('\n')
		if logLine != "" {
			count++
		} else {
			if io.EOF == err {
				break
			}
		}
	}
	return count
}

// get last 10 min log file name
func GetLast10MinLogFile() string {
	currentTime := time.Now().Unix()
	last10Min := currentTime - currentTime%300 - 600
	fileName := conf.Conf.Log.TargetPath + conf.Conf.Log.TargetFilePrefix + "-" + time.Unix(last10Min, 0).Format("2006-01-02-15-04") + ".log"
	return fileName
}
