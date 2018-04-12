package log

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

//default channel buf length is 1k
const defaultAsyncMsgLen = 1024

type LoggerConfig struct {
	Level       uint8
	CallDepth   uint8
	EnableDepth bool
	IsAsync     bool
	MsgChanLen  uint32
}

func NewLogger(cfg *LoggerConfig, store Store, layout Layout) *Logger {
	l := new(Logger)
	l.level = cfg.Level
	l.enableCall = cfg.EnableDepth
	l.callDepth = cfg.CallDepth
	l.msgChanLen = cfg.MsgChanLen
	l.signalChan = make(chan string, 1)
	l.Store = store
	l.Layout = layout
	return l
}

// set default logger
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

type Logger struct {
	sync.Mutex
	level      uint8
	init       bool
	enableCall bool
	callDepth  uint8
	isAsync    bool
	msgChanLen uint32
	msgChan    chan string // data channel
	signalChan chan string // control channel
	wg         sync.WaitGroup
	Layout     Layout
	Store      Store
}

// set the log to asynchronous and start the goroutine
func (l *Logger) Async(msgLen uint32) *Logger {
	l.Lock()
	defer l.Unlock()
	if l.isAsync {
		return l
	}
	l.isAsync = true
	l.msgChanLen = msgLen
	l.msgChan = make(chan string, l.msgChanLen)
	l.wg.Add(1)
	go l.run()
	return l
}

func (l *Logger) store(s *string) (err error) {
	if len(*s) == 0 {
		return
	}

	if l.Store == nil {
		return fmt.Errorf("Store empty")
	}

	if !l.init {
		l.init = true
		l.Store.Init()
	}

	err = l.Store.WriteMsg(s)
	return err
}

func (l *Logger) writeMsg(level uint8, format interface{}, v ...interface{}) (err error) {
	out := new(string)
	encoded := l.formatLog(format, v...)
	if l.Layout == nil {
		err = fmt.Errorf("lauout nil")
		l.checkErr(err)
		return
	}

	t := time.Now()
	if l.enableCall {
		//get file and line number
		_, file, line, ok := runtime.Caller(int(l.callDepth))
		if !ok {
			file = "???"
			line = 0
		}
		out = l.Layout.Layout(&LayoutInfo{
			Level:      level,
			Msg:        encoded,
			Time:       &t,
			EnableCall: true,
			FileName:   &file,
			LineNumber: line,
		})
	} else {
		out = l.Layout.Layout(&LayoutInfo{
			Level:      level,
			Msg:        encoded,
			Time:       &t,
			EnableCall: false,
		})
	}

	if l.isAsync {
		l.msgChan <- *out
	} else {
		err = l.store(out)
		if err != nil {

		}
		return
	}
	return
}

func (l *Logger) formatLog(f interface{}, v ...interface{}) *string {
	var format string
	switch f.(type) {
	case string:
		format = f.(string)
		if len(v) == 0 {
			return &format
		}
		if strings.Contains(format, "%") && !strings.Contains(format, "%%") {

		} else {
			// do not contain format char
			// add %v format expression automatically. important!
			format += strings.Repeat(" %v", len(v))
		}
	default:
		format = fmt.Sprint(f)
		if len(v) == 0 {
			return &format
		}
		format += strings.Repeat(" %v", len(v))
	}
	out := fmt.Sprintf(format, v...)
	return &out
}

func (l *Logger) SetLevel(lvl uint8) {
	l.level = lvl
}

func (l *Logger) SetCallDepth(d uint8) {
	l.callDepth = d
}

func (l *Logger) GetCallDepth() uint8 {
	return l.callDepth
}

func (l *Logger) EnableCall(b bool) {
	l.enableCall = b
}

// if logger get error, dump error msg to stdout
func (l *Logger) checkErr(err error) {
	errMsg := err.Error()
	if len(err.Error()) > 0 {
		fmt.Println("[***** Logger Error *****]:", errMsg)
	}
}

// run for async channel msg
func (l *Logger) run() {
	end := false
	for {
		select {
		case str := <-l.msgChan:
			l.store(&str)

		case sigal := <-l.signalChan:
			// Now should only send "flush" or "close" to l.signalChan
			l.Flush()
			if sigal == "close" {
				l.Store.Destroy()
				end = true
			}
		}
		if end {
			l.wg.Done()
			break
		}
	}
}

// Emergency Log EMERGENCY level message.
func (l *Logger) Emergency(format interface{}, v ...interface{}) {
	if LevelEmergency > l.level {
		return
	}
	l.writeMsg(LevelEmergency, format, v...)
}

// Alert Log ALERT level message.
func (l *Logger) Alert(format interface{}, v ...interface{}) {
	if LevelAlert > l.level {
		return
	}
	l.writeMsg(LevelAlert, format, v...)
}

func (l *Logger) Critical(format interface{}, v ...interface{}) {
	if LevelCritical > l.level {
		return
	}
	l.writeMsg(LevelCritical, format, v...)
}

func (l *Logger) Error(format interface{}, v ...interface{}) {
	if LevelError > l.level {
		return
	}
	l.writeMsg(LevelError, format, v...)
}

func (l *Logger) Notice(format interface{}, v ...interface{}) {
	if LevelNotice > l.level {
		return
	}
	l.writeMsg(LevelNotice, format, v...)
}

func (l *Logger) Debug(format interface{}, v ...interface{}) {
	if LevelDebug > l.level {
		return
	}
	l.writeMsg(LevelDebug, format, v...)
}

func (l *Logger) Warn(format interface{}, v ...interface{}) {
	if LevelWarning > l.level {
		return
	}
	l.writeMsg(LevelWarning, format, v...)
}

func (l *Logger) Info(format interface{}, v ...interface{}) {
	if LevelInformational > l.level {
		return
	}
	l.writeMsg(LevelInformational, format, v...)
}

// flush all data.
func (l *Logger) Flush() {
	if l.isAsync {
		l.signalChan <- "flush"
	}
	l.Store.Flush()
}

// Close close logger, flush all chan data and destroy all adapters in Logger.
func (l *Logger) Close() {
	if l.isAsync {
		l.signalChan <- "close"
		close(l.msgChan)
	}
	l.Store.Flush()
	close(l.signalChan)
}

var defaultLogger = NewLogger(defaultConfig, defaultStore, &BaseLayout{})
var defaultConfig = &LoggerConfig{Level: LevelDebug}
var defaultStore = GetDefaultConsoleStore()

func SetLevel(l uint8) {
	defaultLogger.SetLevel(l)
}

func SetLogFuncCall(b bool) {
	defaultLogger.EnableCall(b)
	defaultLogger.SetCallDepth(3)
}

func EnableAsync() {
	defaultLogger.Async(defaultAsyncMsgLen)
}

func Emergency(f interface{}, v ...interface{}) {
	defaultLogger.Emergency(f, v...)
}

func Alert(f interface{}, v ...interface{}) {
	defaultLogger.Alert(f, v...)
}

func Critical(f interface{}, v ...interface{}) {
	defaultLogger.Critical(f, v...)
}

func Error(f interface{}, v ...interface{}) {
	defaultLogger.Error(f, v...)
}

func Warn(f interface{}, v ...interface{}) {
	defaultLogger.Warn(f, v...)
}

func Notice(f interface{}, v ...interface{}) {
	defaultLogger.Notice(f, v...)
}

func Info(f interface{}, v ...interface{}) {
	defaultLogger.Info(f, v...)
}

func Debug(f interface{}, v ...interface{}) {
	defaultLogger.Debug(f, v...)
}

func Flush() {
	defaultLogger.Flush()
}
