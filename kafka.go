package log

import (
	"fmt"
	"github.com/Shopify/sarama"
	"sync"
)

type KafkaStore struct {
	sync.RWMutex // write log order by order
	Producer     sarama.SyncProducer
	Topic        string
}

func NewKafkaStore(addrs []string, topic string, config *sarama.Config) (Store, error) {
	//p, err := sarama.NewAsyncProducer(addrs, config)
	p, err := sarama.NewSyncProducer(addrs, config)

	if err != nil {
		return nil, err
	}
	producer := new(KafkaStore)
	producer.Producer = p
	producer.Topic = topic
	return producer, nil
}

func (w *KafkaStore) Init() error {
	return nil
}

func (k *KafkaStore) checkErr(err error) {
	errMsg := err.Error()
	if len(err.Error()) > 0 {
		fmt.Println("[***** Kafka Error *****]:", errMsg)
	}

}

// WriteMsg write logger message to kafka.
// if low latency is needed, consider use AsyncProducer
func (k *KafkaStore) WriteMsg(s *string) error {
	k.Lock()
	// check last byte if "\n"
	data := []byte(*s)
	if data[len(data)-1] == 10 {
		data = data[:len(data)-1]
	}
	encodeMsg := &sarama.ProducerMessage{
		Topic: k.Topic,
		Value: sarama.StringEncoder(string(data))}
	//k.AsyncProducer.Input() <- encodeMsg
	k.Producer.SendMessage(encodeMsg)
	k.Unlock()
	return nil
}

func (k *KafkaStore) Destroy() {
	k.Producer.Close()
}

// flush file means sync file from disk.
func (k *KafkaStore) Flush() {
}
