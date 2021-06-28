package store

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type FileStore struct {
	locker  *sync.RWMutex
	dirname string
	files   map[string]*os.File
	caps    map[string]int
}

func NewFileStore(dirname string) Store {
	return &FileStore{
		locker:  new(sync.RWMutex),
		dirname: dirname,
		files:   make(map[string]*os.File),
		caps:    make(map[string]int),
	}
}

func (fs *FileStore) Open(topic string) {
	fs.locker.Lock()
	defer fs.locker.Unlock()
	logName := fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
	file, err := os.OpenFile(logName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err.Error())
	}
	fs.files[topic] = file

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

// ReadAll 读取主题下所有消息，日志分片，需要读取多个文件
func (fs *FileStore) ReadAll(topic string) []common.MessageUnit {
	var msgList []common.MessageUnit
	files, _ := ioutil.ReadDir(fs.dirname)
	for _, file := range files{
		match, _ := regexp.MatchString(fmt.Sprintf("gomq.%s.log[.]?[0-9]*", topic), file.Name())
		if match {
			bytes, err := ioutil.ReadFile(fs.dirname + file.Name())
			if err != nil {
				return msgList
			}
			byteList := strings.Split(string(bytes), "\n")
			for _, byteItem := range byteList[:len(byteList)-1] {
				msg := new(common.MessageUnit)
				msgList = append(msgList, *msg.UnPack([]byte(byteItem)))
			}
		}
	}
	return msgList
}

func (fs *FileStore) Close() {
	fs.locker.RLock()
	defer fs.locker.RUnlock()
	for _, file := range fs.files {
		file.Close()
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

func (fs *FileStore) GetAllTopics() (topics []string) {
	for topic, _ := range fs.files {
		topics = append(topics, topic)
	}
	return
}

// GetCompactFd 获取需要进行压缩的文件描述符
func (fs *FileStore) GetCompactFd(topic string) []*os.File {
	files, _ := ioutil.ReadDir(fs.dirname)
	var fds []*os.File
	for _, file := range files {
		if file.Name() != fs.CurrentLogName(topic) {
			fd, err := os.OpenFile(file.Name(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				log.Fatalf("获取待压缩日志失败", err)
			}
			fds = append(fds, fd)
		}
	}
	return fds
}

// GetTargetFilename 获取主题下的将要重命名的文件名
func (fs *FileStore) GetTargetFilename(topic string) string {
	files, _ := ioutil.ReadDir(fs.dirname)
	var filename,targetName string
	for _, file := range files {
		match, _ := regexp.MatchString(fmt.Sprintf("gomq.%s.log[.]?[0-9]*", topic), file.Name())
		if match {
			filename = file.Name()
		}
	}
	split := strings.Split(filename, ".")
	if len(split) == 3 {
		targetName =  filename + ".1"
	} else {
		lastNo, _ := strconv.Atoi(split[3])
		targetName =  fmt.Sprintf("%s.%d", filename, lastNo+1)
	}
	return fs.dirname + targetName
}

// IsNeedSplit 是否需要分片，大于2M则需要
func (fs *FileStore) IsNeedSplit(topic string) bool {
	fd, _ := os.OpenFile(fs.CurrentLogName(topic), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	fi, _ := fd.Stat()
	return fi.Size() > 1024*1024*2
}

func (fs *FileStore) CurrentLogName(topic string) string {
	return fmt.Sprintf("%sgomq.%s.log", fs.dirname, topic)
}
