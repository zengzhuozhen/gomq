package store

import (
	"bufio"
	"fmt"
	"gomq/common"
	"os"
	"sync"
)

type filestore struct {
	locker   *sync.RWMutex
	isOpen   bool
	fileName string
	file     *os.File
	cap      int
}

func NewFileStore(savePath string) Store {
	return &filestore{
		locker:   new(sync.RWMutex),
		isOpen:   false,
		fileName: savePath,
	}
}

func (f *filestore) Open() {
	f.locker.Lock()
	defer f.locker.Unlock()
	if f.isOpen == false {
		file, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Print(err.Error())
			panic("couldn't open the file ")
		}
		f.file = file
	}
}

func (f *filestore) Append(item common.MessageUnit) {
	data := item.Data.Pack()
	data = append(data, []byte("\n")...)
	write := bufio.NewWriter(f.file)
	_, err := write.Write(data)
	if err != nil {
		panic(err.Error())
	}
	err = write.Flush()
	f.cap++
	if err != nil {
		panic(err.Error())
	}
}


func (f *filestore) Reset() {
	return
}

func (f *filestore) Load() {
	return
}

func (f *filestore) Close() {
	f.locker.RLock()
	defer f.locker.RUnlock()
	if f.isOpen == true {
		f.file.Close()
		f.isOpen = false
	}
	return
}

func (f *filestore) Cap() int {
	return f.cap
}
