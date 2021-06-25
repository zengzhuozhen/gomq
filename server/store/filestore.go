package store

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type FileStore struct {
	locker  *sync.RWMutex
	dirname string
	isOpen  map[string]bool
	files   map[string]*os.File
	caps    map[string]int
}

func NewFileStore(dirname string) Store {
	return &FileStore{
		locker:  new(sync.RWMutex),
		dirname: dirname,
		isOpen:  make(map[string]bool),
		files:   make(map[string]*os.File),
		caps:    make(map[string]int),
	}
}

func (fs *FileStore) Open(topic string) {
	fs.locker.Lock()
	defer fs.locker.Unlock()
	if fs.isOpen[topic] == false {
		logName := fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
		file, err := os.OpenFile(logName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err.Error())
		}
		fs.files[topic] = file
		fs.isOpen[topic] = true
	}
}

func (fs *FileStore) Append(item common.MessageUnit) {
	data := item.Pack()
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

func (fs *FileStore) Reset(topic string) {
	logName := fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
	err := ioutil.WriteFile(logName, []byte{}, os.ModePerm)
	if err != nil {
		panic(err.Error())
	}
}

func (fs *FileStore) ReadAll(topic string) []common.MessageUnit {
	var msgList []common.MessageUnit
	logName := fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
	bytes, err := ioutil.ReadFile(logName)
	if err != nil {
		return msgList
	}
	byteList := strings.Split(string(bytes), "\n")
	for _, byteItem := range byteList[:len(byteList)-1] {
		msg := new(common.MessageUnit)
		msgList = append(msgList, *msg.UnPack([]byte(byteItem)))
	}
	return msgList
}

func (fs *FileStore) Close() {
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

func (fs *FileStore) Cap(topic string) int {
	logName := fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
	buf, err := ioutil.ReadFile(logName)
	fmt.Println(bytes.Count(buf, []byte{'\n'}))
	if err != nil {
		return 0
	}
	return bytes.Count(buf, []byte{'\n'})
}


func (fs *FileStore)GetAllTopics () (topics []string){
	for topic , _ := range fs.files{
		topics = append(topics, topic)
	}
	return
}

func (fs *FileStore) GetFd(topic string) *os.File {
	logName := fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
	fd , err := os.OpenFile(logName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil{
		panic(err)
	}
	return fd
}