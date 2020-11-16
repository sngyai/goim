package dao

import (
	"errors"
	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/nats-io/nats.go"
)


func newNats(c *conf.Nats) *nats.Conn {
	nc, err := nats.Connect(c.Broker)
	if err != nil {
		panic(err)
	}
	return nc
}

func newNatsDao(c *conf.Config) *natsDao{
	dao := &natsDao{
		c: c,
		push: newNats(c.Nats),
	}
	return dao
}


func (d *natsDao) PublishMessage(c *conf.Config, key string, value []byte) error {
	if d.push == nil {
		return errors.New("nats error")
	}
	return d.push.Publish(c.Nats.Topic, value)

}

func (d *natsDao) Close() error{
	d.push.Close()
	return nil
}

