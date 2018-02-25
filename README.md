## Log

Easy and userful in golang




***
## Usage

Get codes  
` $ go get github.com/miaomiao3/log`

Import  
`import ( "github.com/miaomiao3/log" )`

***

Examples

### console and file
package main

import (
	"github.com/miaomiao3/log"
)

func main() {

	// default is console
	log.Debug("debug")
	log.Info("informational")
	log.Notice("notice")
	log.Warn("warning")
	log.Error("error")
	log.Critical("critical")
	log.Alert("alert")
	log.Emergency("emergence")

	// file option
	fileCfg := &log.FileConfig{
		FileName:    "test.log",
		MaxDays:     3,    //delete the old file after 7 days
		MaxSize:     100,  //rename the old file when its lines > Maxlines
		DailyRotate: true, //rename the old file when date changes and start a new log file
	}

	fileStore, err := log.NewFileStore(fileCfg)
	if err != nil {
		panic(err)
	}

	loggerCfg := &log.LoggerConfig{
		Level:       log.LevelDebug, // emit when priority lower than debug
		CallDepth:   2,
		EnableDepth: true,
		IsAsync:     false,
	}

	logger := log.NewLogger(loggerCfg, fileStore, &log.BaseLayout{})
	for i := 0; i < 50; i++ {
		logger.Debug("testing %v", i)
	}

}


```
