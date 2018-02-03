package log

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestFileStore(t *testing.T) {
	Convey("fileStore", t, func() {
		fileCfg := &FileConfig{
			FileName:    "test.log",
			MaxDays:     3,    //delete the old file after 7 days
			MaxSize:     100,  //rename the old file when its lines > Maxlines
			DailyRotate: true, //rename the old file when date changes and start a new log file
		}
		fileStore, err := NewFileStore(fileCfg)

		So(err, ShouldEqual, nil)
		fileStore.Init()
		msg := "123"
		fileStore.WriteMsg(&msg)
		os.Remove("test.log")
	})

}
