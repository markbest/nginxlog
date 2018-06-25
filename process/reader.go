package process

import (
	"bufio"
	"io"
	"os"
)

type ReadFromFile struct {
	FilePath string
}

// read from file
func (r *ReadFromFile) Read(readChan chan string) {
	file, err := os.Open(r.FilePath)
	if err != nil {
		return
	}
	defer file.Close()

	buff := bufio.NewReader(file)
	for {
		logLine, err := buff.ReadString('\n')
		if logLine != "" {
			readChan <- logLine
		} else {
			if io.EOF == err {
				break
			}
		}
	}
	close(readChan)
}
