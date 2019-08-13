package cmd

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hpcloud/tail"
)

type tailer struct {
	tailer        *tail.Tail
	parser        func(string) (string, error)
	offsetFile    *os.File
	logFile       string
	configuration tail.Config
}

func NewHPTailer(parser Parser, logFileToTail string) (*tailer, error) {
	t := &tailer{}
	var err error

	t.parser = Parse

	//Calculate offset for log file
	curOffset, fp, err := fetchOffset(logFileToTail)
	if err != nil {
		log.Fatalf("[ERROR] %+v", err)
		return nil, err
	}
	t.offsetFile = fp
	si := tail.SeekInfo{
		Offset: int64(curOffset),
	}
	t.configuration = tail.Config{Follow: true, Location: &si}
	t.logFile = logFileToTail
	t.tailer, err = tail.TailFile(t.logFile, t.configuration)
	if err != nil {
		log.Fatal(err)
	}
	t.logFile = logFileToTail
	return t, nil
}

func (myT *tailer) Start() {
	defer myT.Cleanup()
	var err error
	done := make(chan bool)
	//!!Update
	go func(t *tail.Tail) {
		for {
			select {
			case event := <-t.Lines:
				if event == nil {
					log.Printf("[INFO] File rotated")
					updateOffset(myT.tailer, myT.offsetFile) //Truncate
					time.Sleep(3000)
					myT.tailer, err = tail.TailFile(myT.logFile, myT.configuration)
					if err != nil {
						log.Fatal(err)
					}
				} else if event.Text != "" {
					r, err := myT.parser(event.Text)
					if err != nil || r == "" {
						if err != nil {
							log.Printf("[ERROR] %+v", err)
						}
					} else {
						err := Send(r) //Send message to the stash
						if err != nil {
							log.Printf("[ERROR] %+v", err)
						}
					}
					updateOffset(myT.tailer, myT.offsetFile) //Truncate

				}
			}

		}
	}(myT.tailer)

	<-done
}

func (myT *tailer) Cleanup() {
	myT.tailer.Cleanup()
}

func updateOffset(t *tail.Tail, f *os.File) {
	if t != nil {
		f.Truncate(0)
		offset, err := t.Tell()
		if err != nil {
			println("Error:" + err.Error())
		}
		f.WriteString(strconv.Itoa(int(offset)))
	} else {
		f.Truncate(0)
	}
}

func fetchOffset(pathToLog string) (int, *os.File, error) {
	//Create offsetfile name by using the path to the file
	h := md5.Sum([]byte(pathToLog))
	fn := fmt.Sprintf("%x", h) + ".offset"

	//Stat offset file, if it does not exist create an empty file
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		os.Create(fn)
	}

	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return -1, nil, err
	}
	var curOffset int
	if len(b) == 0 {
		curOffset = 0
	} else {
		curOffset, err = strconv.Atoi(string(b[:]))
		if err != nil {
			return -1, nil, err
		}
	}

	fp, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return curOffset, fp, nil
}
