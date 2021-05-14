package visit

import (
	"github.com/zengzhuozhen/gomq/common"
)

// container for visitor,it provide some struct and invoke function in visit func
type container struct {
	retainQueue *common.RetainQueue
	connection *common.ConnectionAbstract
}