package process

import (
	"encoding/json"
	"github.com/markbest/nginxlog/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type reader interface {
	Read(readChan chan string)
}

type writer interface {
	Write(writeChan chan string)
}

// log process struct
type LogProcess struct {
	ReadChan  chan string
	WriteChan chan string
}

// read source
func (l *LogProcess) ReadSource(r reader) {
	r.Read(l.ReadChan)
}

// parse log data
func (l *LogProcess) ParseLogData() {
	for {
		data, ok := <-l.ReadChan
		if !ok {
			break
		}

		// remote addr
		remoteAddrReg := regexp.MustCompile(`([0-9]{1,3}\.){3}[0-9]{1,3}`)
		remoteAddrData := remoteAddrReg.FindAllString(data, -1)[0]

		// remote user
		remoteUserData := ""

		// time local
		timeLocalReg := regexp.MustCompile(`\[\d{1,2}\/\w{3}\/\d{1,4}(:[0-9]{1,2}){3} \+([0-9]){1,4}\]`)
		timeLocalData := timeLocalReg.FindAllString(data, -1)[0]
		parsedTime, _ := time.Parse("[02/Jan/2006:15:04:05 -0700]", timeLocalData)
		parsedTimeUnix := parsedTime.Unix()

		// request type
		requestTypeReg := regexp.MustCompile(`"\w+`)
		requestTypeData := requestTypeReg.FindAllString(data, -1)[0]
		requestTypeData = requestTypeData[1:]

		// request url
		requestUrlReg := regexp.MustCompile(`"\w+\s[^\s]+`)
		requestUrlData := requestUrlReg.FindAllString(data, -1)[0]
		requestUrlData = requestUrlData[5:]

		// http version
		httpVersionReg := regexp.MustCompile(`HTTP\/\d.\d"`)
		httpVersionData := httpVersionReg.FindAllString(data, -1)[0]
		httpVersionData = httpVersionData[:len(httpVersionData)-1]

		// status && body bytes sent
		responseAndByteReg := regexp.MustCompile(`([0-9]{1,3}) \d+`)
		responseAndByteData := responseAndByteReg.FindAllString(data, -1)[0]
		str := strings.Split(responseAndByteData, " ")
		statusData, _ := strconv.Atoi(str[0])
		bodyBytesSentData, _ := strconv.Atoi(str[1])

		// http referer
		var httpReferer string
		httpRefererReg := regexp.MustCompile(`(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`)
		httpRefererData := httpRefererReg.FindAllString(data, -1)
		if len(httpRefererData) > 0 {
			httpReferer = httpRefererData[0]
		} else {
			httpReferer = ""
		}

		// http user agent
		str = strings.Split(data, "\" \"")
		httpUserAgentData := str[len(str)-1]
		httpUserAgentData = httpUserAgentData[0 : len(httpUserAgentData)-2]

		// http x forwarded for data
		httpXForwardedForData := ""

		// append parse result to content
		rs := &utils.LogFormat{
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
		jsonRs, _ := json.Marshal(rs)
		l.WriteChan <- string(jsonRs)
	}
}

// write to target
func (l *LogProcess) WriteTarget(w writer) {
	w.Write(l.WriteChan)
}
