package log

import (
	"testing"
)

func TestConsole(t *testing.T) {
	store:=GetDefaultConsoleStore()
	store.Init()
	msg := "123"
	store.WriteMsg(&msg)
}