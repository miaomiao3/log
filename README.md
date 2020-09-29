## Log

Easy logger




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
	"github.com/Shopify/sarama"
	"github.com/miaomiao3/log"
	"os"
	"os/signal"
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

	// file store
	fileCfg := &log.FileConfig{
		FileName:    "test.log",
		MaxDays:     3,    //delete the old file after 7 days
		MaxSize:     100,  //rename the old file when its size larger than MaxSize
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
	for i := 0; i < 10; i++ {
		logger.Debug("testing %v", i)
	}

	closeSig := make(chan os.Signal, 1)
	signal.Notify(closeSig, os.Interrupt)

	<-closeSig

	log.Debug("exit")

}


```
