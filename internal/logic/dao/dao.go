package dao

import (
	"context"
	pb "github.com/Terry-Mao/goim/api/logic/grpc"
	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"github.com/nats-io/nats.go"
	"strconv"
	"time"

	"github.com/Terry-Mao/goim/internal/logic/conf"
	"github.com/gomodule/redigo/redis"
	kafka "gopkg.in/Shopify/sarama.v1"
)

// Dao dao.
type Dao struct {
	c           *conf.Config
	//kafkaPub    kafka.SyncProducer
	push        PushMsg
	//nats		*nats.Conn
	redis       *redis.Pool
	redisExpire int32
}

// natsDao dao for nats
type natsDao struct {
	c    *conf.Config
	push *nats.Conn
}

type kafkaDao struct {
	c	*conf.Config
	push kafka.SyncProducer
}

type PushMsg interface {
	PublishMessage(c *conf.Config, key string, msg []byte) error // ****** 这里小改了个方法名!!! 注意
	Close() error
}

// New new a dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		//kafkaPub:    newKafkaPub(c.Kafka),
		redis:       newRedis(c.Redis),
		redisExpire: int32(time.Duration(c.Redis.Expire) / time.Second),
	}
	log.Errorf("Dao config: %v", c)
	switch c.Mq {
	case "nats":
		d.push = newNatsDao(c)
	case "kafka":
		d.push = newKafkaDao(c)
	default:
		panic("invalid config")
	}

	return d
}

func newRedis(c *conf.Redis) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.Idle,
		MaxActive:   c.Active,
		IdleTimeout: time.Duration(c.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(c.Network, c.Addr,
				redis.DialConnectTimeout(time.Duration(c.DialTimeout)),
				redis.DialReadTimeout(time.Duration(c.ReadTimeout)),
				redis.DialWriteTimeout(time.Duration(c.WriteTimeout)),
				redis.DialPassword(c.Auth),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

// Close close the resource.
func (d *Dao) Close() error {
	return d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}

// PushMsg push a message to databus.
func (d *Dao) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	if err = d.push.PublishMessage(d.c, keys[0], b); err != nil {
		log.Errorf("PushMsg.send(push pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// BroadcastRoomMsg push a message to databus.
func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, room string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Room:      room,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	if err = d.push.PublishMessage(d.c, room, b); err != nil {
		log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// BroadcastMsg push a message to databus.
func (d *Dao) BroadcastMsg(c context.Context, op, speed int32, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Speed:     speed,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	if err = d.push.PublishMessage(d.c, strconv.FormatInt(int64(op), 10), b); err != nil {
		log.Errorf("PushMsg.send(broadcast pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}
