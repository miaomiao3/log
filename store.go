package log

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
