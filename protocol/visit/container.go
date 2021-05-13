package visit

import (
	"github.com/zengzhuozhen/gomq/common"
	"github.com/zengzhuozhen/gomq/server/service"
)

// container for visitor,it provide some struct and invoke function in visit func
type container struct {
	retainQueue *common.RetainQueue
	connections map[string]*service.ConnectionAbstract
}