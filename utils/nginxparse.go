package utils

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Nginx data fields as per the Split Function
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

// Struct for the object
type LogStreamData struct {
	fileHandle *os.File
	fileErr    error
}

// Reading the file
func (log *LogStreamData) OpenStream(filename string) error {
	currentTime := time.Now().Unix()
	last10min := currentTime - currentTime%300 - 600
	filename = filename + "-" + time.Unix(last10min, 0).Format("2006-01-02-15-04") + ".log"
	log.fileHandle, log.fileErr = os.Open(filename)
	if log.fileErr != nil {
		return log.fileErr
	}
	return nil
}

// Here lines are read and mapped to data struct
func (log *LogStreamData) ParseStream(client *ES, elasticIndex string, elasticType string) {
	maxReadLine := 100
	line := 1
	parseTimeCount := 1
	buff := bufio.NewReader(log.fileHandle)
	content := make([]*LogFormat, 0)
	for {
		logLine, err := buff.ReadString('\n')
		if logLine != "" {
			content = append(content, parseLogContent(logLine))
			if line%maxReadLine == 0 {
				log.output(content[(parseTimeCount-1)*maxReadLine:], client, elasticIndex, elasticType)
				parseTimeCount++
			}
		} else {
			if io.EOF == err {
				log.output(content[(parseTimeCount-1)*maxReadLine:], client, elasticIndex, elasticType)
				break
			}
		}
		line++
	}
}

// close the file handle
func (log *LogStreamData) CloseStream() {
	log.fileHandle.Close()
}

// output file parsed content
func (log *LogStreamData) output(content []*LogFormat, client *ES, elasticIndex string, elasticType string) {
	if len(content) > 0 {
		for _, v := range content {
			_, err := client.EsClient.Index().
				Index(elasticIndex).
				Type(elasticType).
				BodyJson(*v).
				Do()
			if err != nil {
				panic(err)
			}
		}
	}
}

// Regexp parse log line content
func parseLogContent(logLine string) (rs *LogFormat) {
	// remote addr
	remoteAddrReg := regexp.MustCompile(`([0-9]{1,3}\.){3}[0-9]{1,3}`)
	remoteAddrData := remoteAddrReg.FindAllString(logLine, -1)[0]

	// remote user
	remoteUserData := ""

	// time local
	timeLocalReg := regexp.MustCompile(`\[\d{1,2}\/\w{3}\/\d{1,4}(:[0-9]{1,2}){3} \+([0-9]){1,4}\]`)
	timeLocalData := timeLocalReg.FindAllString(logLine, -1)[0]
	parsedTime, _ := time.Parse("[02/Jan/2006:15:04:05 -0700]", timeLocalData)
	parsedTimeUnix := parsedTime.Unix()

	// request type
	requestTypeReg := regexp.MustCompile(`"\w+`)
	requestTypeData := requestTypeReg.FindAllString(logLine, -1)[0]
	requestTypeData = requestTypeData[1:]

	// request url
	requestUrlReg := regexp.MustCompile(`"\w+\s[^\s]+`)
	requestUrlData := requestUrlReg.FindAllString(logLine, -1)[0]
	requestUrlData = requestUrlData[5:]

	// http version
	httpVersionReg := regexp.MustCompile(`HTTP\/\d.\d"`)
	httpVersionData := httpVersionReg.FindAllString(logLine, -1)[0]
	httpVersionData = httpVersionData[:len(httpVersionData)-1]

	// status && body bytes sent
	responseAndByteReg := regexp.MustCompile(`([0-9]{1,3}) \d+`)
	responseAndByteData := responseAndByteReg.FindAllString(logLine, -1)[0]
	str := strings.Split(responseAndByteData, " ")
	statusData, _ := strconv.Atoi(str[0])
	bodyBytesSentData, _ := strconv.Atoi(str[1])

	// http referer
	var httpReferer string
	httpRefererReg := regexp.MustCompile(`(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`)
	httpRefererData := httpRefererReg.FindAllString(logLine, -1)
	if len(httpRefererData) > 0 {
		httpReferer = httpRefererData[0]
	} else {
		httpReferer = ""
	}

	// http user agent
	str = strings.Split(logLine, "\" \"")
	httpUserAgentData := str[len(str)-1]
	httpUserAgentData = httpUserAgentData[0 : len(httpUserAgentData)-2]

	// http x forwarded for data
	httpXForwardedForData := ""

	// append parse result to content
	rs = &LogFormat{
		RemoteAddr:        remoteAddrData,
		RemoteUser:        remoteUserData,
		TimeLocal:         parsedTime,
		RequestType:       requestTypeData,
		RequestUrl:        requestUrlData,
		HttpVersion:       httpVersionData,
		Status:            statusData,
		BodyBytesSent:     bodyBytesSentData,
		HttpReferer:       httpReferer,
		HttpUserAgent:     httpUserAgentData,
		HttpXForwardedFor: httpXForwardedForData,
		CreatedAt:         parsedTimeUnix,
	}
	return rs
}
