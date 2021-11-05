// Package dumperror Output terminal to file
// and
package dumperror

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var l = logger.DefaultSLogger("dump")

func StartDump() {
	Start()
}
