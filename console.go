package log

import (
	"os"
	"sync"
	"io"
)


type ConsoleStore struct {
	sync.Mutex
	writer     io.Writer
}

func DefaultConsoleStore() Store {
	console := &ConsoleStore{
		writer:     os.Stdout,
	}
	return console
}

func (c *ConsoleStore) Init() error {
	return nil
}

func (c *ConsoleStore) WriteMsg(s *string) error {
	c.Lock()
	defer c.Unlock()
	c.writer.Write([]byte(*s))
	return nil
}

// Destroy implementing method. empty.
func (c *ConsoleStore) Destroy() {

}

// Flush implementing method. empty.
func (c *ConsoleStore) Flush() {

}