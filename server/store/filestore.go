package store

import (
	"bufio"
	"fmt"
	"gomq/common"
	"os"
	"sync"
)

type filestore struct {
	locker   sync.RWMutex
	isOpen   bool
	fileName string
	file     *os.File
	data     []common.MessageUnit
}

func NewFileStore() Store {
	return &filestore{
		locker:   sync.RWMutex{},
		isOpen:   false,
		fileName: "/var/log/tempmq.log",
		data:     make([]common.MessageUnit,0),
	}
}

func (f *filestore) Open() {
	f.locker.Lock()
	defer f.locker.Unlock()
	if f.isOpen == true{
		return
	}else{
		file , err := os.OpenFile(f.fileName,os.O_WRONLY|os.O_APPEND,0666)
		if err != nil{
			fmt.Print(err.Error())
			panic("couldn't open the file ")
		}
		f.file = file
	}
}

func (f *filestore) Append(item common.MessageUnit) {
	f.data = append(f.data,item)
	data := item.Data.Pack()
	write := bufio.NewWriter(f.file)
	_,err := write.Write(data)
	if err != nil{
		panic(err.Error())
	}
	err = write.Flush()
	if err != nil{
		panic(err.Error())
	}
}

func (f *filestore) SnapShot() {
	return
}

func (f *filestore) Reset() {
	return
}

func (f *filestore) Load() {
	return
}

func (f *filestore) Close() {
	f.file.Close()
}

func (f *filestore) Cap() int {
	return len(f.data)
}
