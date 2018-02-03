package log

import (
	"fmt"
)

// store
const (
	StoreConsole = 1
	StoreFile    = 2
	StoreKafka   = 3
)

type Store interface {
	Init() error
	WriteMsg(s *string) error
	Flush()
	Destroy()
}

var StoreMap = make(map[int]*Store)

func RegisterStore(storeId int, store *Store) {
	if store == nil {
		fmt.Println("store empty")
	}

	if _, ok := StoreMap[storeId]; !ok {
		StoreMap[storeId] = store
	}
}
