package log

import (
	"testing"
)

func TestConsole(t *testing.T) {
	store:=DefaultConsoleStore()
	store.Init()
	msg := "123"
	store.WriteMsg(&msg)
}