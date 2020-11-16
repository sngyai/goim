package dao

import (
	"errors"
	"github.com/Shopify/sarama"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	kafka "gopkg.in/Shopify/sarama.v1"
)

func newKafkaDao(c *conf.Config) *kafkaDao{
	dao := &kafkaDao{
		c: c,
		push: newKafkaPub(c.Kafka),
	}
	return dao
}

func newKafkaPub(c *conf.Kafka) kafka.SyncProducer {
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll // Wait for all in-sync replicas to ack the message
	kc.Producer.Retry.Max = 10                  // Retry up to 10 times to produce the message
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

func (d *kafkaDao) PublishMessage(c *conf.Config, key string, value []byte) error {
	if d.push == nil{
		return errors.New("kafka error")
	}
	m := &kafka.ProducerMessage{
		Key:   sarama.StringEncoder(key),
		Topic: c.Kafka.Topic,
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := d.push.SendMessage(m)
	return err
}

func(d *kafkaDao) Close() error{
	d.push.Close()
	return nil
}

