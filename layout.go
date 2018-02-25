package log

import (
	"strconv"
	"time"
)

var defaultLevelPrefix = [LevelDebug + 1]string{"[M] ", "[A] ", "[C] ", "[E] ", "[W] ", "[N] ", "[I] ", "[D] "}

type LayoutInfo struct {
	Level      uint8
	Msg        *string
	Time          *time.Time
	EnableCall bool
	FileName   *string // file name via caller, related to callDepth
	LineNumber int     // code line number, related to callDepth
}

// layout interface
type Layout interface {
	Layout(info *LayoutInfo) *string
}


type BaseLayout struct{}

func (l *BaseLayout) Layout(info *LayoutInfo) *string {
	out := ""
	if info.Level < 8 {
		out = defaultLevelPrefix[info.Level] + *info.Msg
	}

	if info.EnableCall {
		out = "[" + *info.FileName + ":" + strconv.Itoa(info.LineNumber) + "] " + out
	}

	out = info.Time.Format("2006/01/02 15:04:05.00 ") + out + "\n"
	return &out
}
