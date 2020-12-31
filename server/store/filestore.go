package store

import (
	"bufio"
	"bytes"
	"fmt"
	"gomq/common"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type filestore struct {
	locker  *sync.RWMutex
	dirname string
	isOpen  map[string]bool
	files   map[string]*os.File
	caps    map[string]int
}

func NewFileStore(dirname string) Store {
	return &filestore{
		locker:  new(sync.RWMutex),
		dirname: dirname,
		isOpen:  make(map[string]bool),
		files:   make(map[string]*os.File),
		caps:    make(map[string]int),
	}
}

func (fs *filestore) Open(topic string) {
	fs.locker.Lock()
	defer fs.locker.Unlock()
	if fs.isOpen[topic] == false {
		logName := fmt.Sprintf("%s%s.log", fs.dirname, topic)
		file, err := os.OpenFile(logName, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil && os.IsNotExist(err) {
			file, err = os.Create(logName)
		}
		if err != nil {
			panic(err.Error())
		}
		fs.files[topic] = file
	}
}

func (fs *filestore) Append(item common.MessageUnit) {
	data := item.Data.Pack()
	data = append(data, []byte("\n")...)
	write := bufio.NewWriter(fs.files[item.Topic])
	_, err := write.Write(data)
	if err != nil {
		panic(err.Error())
	}
	err = write.Flush()
	fs.caps[item.Topic]++
	if err != nil {
		panic(err.Error())
	}
}

func (fs *filestore) Reset(string) {
	return
}

func (fs *filestore) ReadAll(topic string) []common.MessageUnit {
	var msgList []common.MessageUnit
	if filePtr := fs.files[topic]; filePtr != nil {
		bytes, err := ioutil.ReadAll(filePtr)
		if err != nil {
			return msgList
		}
		byteList := strings.Split(string(bytes), "\n")
		for _, byteItem := range byteList {
			msg := new(common.MessageUnit)
			msgList = append(msgList, *msg.UnPack([]byte(byteItem)))
		}
	}
	return msgList
}

func (fs *filestore) Close() {
	fs.locker.RLock()
	defer fs.locker.RUnlock()
	for topic, file := range fs.files {
		if fs.isOpen[topic] == true {
			file.Close()
			fs.isOpen[topic] = false
		}
	}
	return
}

func (fs *filestore) Cap(topic string) int {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}
	if filePtr := fs.files[topic]; filePtr == nil {
		return 0
	}
	for {
		c, err := fs.files[topic].Read(buf)
		if err != nil && os.IsNotExist(err) {
			return 0
		}
		count += bytes.Count(buf[:c], lineSep)
		if err == io.EOF {
			return count
		}
	}
}
