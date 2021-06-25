package service

import (
	"bufio"
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/log"
	"github.com/zengzhuozhen/gomq/server/store"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"time"
)

type LogManager struct {
	fileStore *store.FileStore
	wg        errgroup.Group
}

func NewLogManager(fileStore *store.FileStore) *LogManager {
	return &LogManager{fileStore: fileStore}
}

func (l *LogManager) Start() error {
	l.wg = errgroup.Group{}
	l.wg.Go(l.startLogCompact)
	l.wg.Go(l.startLogSplit)
	return l.wg.Wait()
}

func (l *LogManager) startLogCompact() error {
	l.inTicker(60*time.Second, func() {
		log.Infof("开启日志压缩")

		topics := l.fileStore.GetAllTopics()
		log.Debugf("压缩日志的topic:", topics)
		for _, topic := range topics { // 相当于重新覆盖，考虑更好的实现
			i := 0
			var dirtyRow []int64
			msgMap := make(map[string]common.MessageUnitWithSort)
			messages := l.fileStore.ReadAll(topic)
			for _, message := range messages {
				if msg, exist := msgMap[message.Data.MsgKey]; exist { // 从前往后扫描，出现重复的key，说明前面的key需要被清除
					dirtyRow = append(dirtyRow, msg.Sort)
				}
				msgMap[message.Data.MsgKey] = common.MessageUnitWithSort{
					Sort:        int64(i),
					MessageUnit: message,
				}
				i++
			}
			for _, fd := range l.fileStore.GetCompactFd(topic) {
				l.compact(fd, dirtyRow)
			}
		}
	})
	return nil
}

// 日志压缩:逐行读取，写入buffer,最后重新覆盖源文件
func (l *LogManager) compact(fd *os.File, dirtyRows []int64) {
	log.Debugf("开始清除日志行", dirtyRows)
	reader := bufio.NewReader(fd)
	res := make([]byte, 0)
	var i int64
	for {
		str, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if !common.FindInt64(i, dirtyRows) {
			res = append(res, []byte(str)...)
		}
		i++
	}
	fd.Truncate(0)
	fd.Write(res)
	fd.Sync()
	fd.Close()
}

// startLogSplit 日志分片
// 参考linux logrotate create
// 1.重命名源文件
// 2.创建新的文件
// 3.重新打开文件
func (l *LogManager) startLogSplit() error {
	l.inTicker(60*time.Second, func() {
		log.Infof("开启日志分片")
		topics := l.fileStore.GetAllTopics()
		for _, topic := range topics {
			if l.fileStore.IsNeedSplit(topic) {
				targetFilename := l.fileStore.GetTargetFilename(topic)
				sourceFilename := l.fileStore.CurrentLogName(topic)
				log.Debugf("源文件:", sourceFilename, "目标文件:", targetFilename)
				_ = os.Rename(sourceFilename, targetFilename)
				_, _ = os.Create(sourceFilename)
				l.fileStore.Open(topic)
			}
		}
	})
	return nil
}

// inTicker let the handler run in  a  new ticker  by given interval
func (l *LogManager) inTicker(interval time.Duration, handler func()) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			handler()
		}
	}
}
