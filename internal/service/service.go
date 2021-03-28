package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bsergik/telebot-nats/internal/database"
	"github.com/bsergik/telebot-nats/internal/telebot"
	"github.com/bsergik/telebot-nats/pkg/external/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

/******************************************************************************
 * Types
 */

type Service struct {
	Zaplog  *zap.Logger
	Storage database.IStorage
	Telebot *telebot.Telebot

	NatsAddress string
	ClusterID   string
	ClientID    string

	MQSubjectName    string
	MQQueueGroupName string
	MQDurableName    string
}

/******************************************************************************
 * Constants
 */

const (
	MQParseDuration    = "60s"
	MQStanPingInterval = 12 // seconds
	MQStanPingMaxOut   = 5
)

/******************************************************************************
 * Global variables
 */

var (
	_gZaplogSugar *zap.SugaredLogger
	_gStorage     database.IStorage
	_gTelebot     *telebot.Telebot
)

/******************************************************************************
 * Методы
 */

func (s *Service) Start(wg *sync.WaitGroup, finish <-chan struct{}) {
	s.initGlobals()

	stanConn := s.stanConnect()
	sub := s.queueSubscribe(stanConn)

	_gZaplogSugar.Info("Started.")

	<-finish

	s.stop(sub)
	wg.Done()
}

func (s *Service) initGlobals() {
	_gZaplogSugar = s.Zaplog.Sugar()
	_gStorage = s.Storage
	_gTelebot = s.Telebot
}

func (s *Service) stanConnect() stan.Conn {
	stanConn, err := stan.Connect(s.ClusterID, s.ClientID,
		stan.NatsURL(s.NatsAddress),
		stan.Pings(MQStanPingInterval, MQStanPingMaxOut),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			_gZaplogSugar.Fatalf("Connection lost, reason: %v", reason)
		}))

	if err != nil {
		_gZaplogSugar.Fatalf("Cannot subscribe chan. Error: %s. Server: %s. "+
			"ClusterID: %s. ClientID: %s. Subject: %s. QGroup: %s. Durable: %s",
			err, s.NatsAddress, s.ClusterID, s.ClientID, s.MQSubjectName, s.MQQueueGroupName, s.MQDurableName)
	}

	return stanConn
}

func (s *Service) queueSubscribe(sc stan.Conn) stan.Subscription {
	aw, _ := time.ParseDuration(MQParseDuration)
	sub, err := sc.QueueSubscribe(
		s.MQSubjectName,
		s.MQQueueGroupName,
		messageProcessor,
		stan.SetManualAckMode(),
		stan.DurableName(s.MQDurableName),
		stan.AckWait(aw),
	)

	if err != nil {
		_gZaplogSugar.Fatalf("Cannot subscribe chan. Error: %s. Server: %s. "+
			"ClusterID: %s. ClientID: %s. Subject: %s. QGroup: %s. Durable: %s",
			err, s.NatsAddress, s.ClusterID, s.ClientID, s.MQSubjectName, s.MQQueueGroupName, s.MQDurableName)
	}

	_gZaplogSugar.Infof("Connected as %q to subject %q (qgroup %q, durable %q) on NATS server %q, cluster-id %q.",
		s.ClientID, s.MQSubjectName, s.MQQueueGroupName, s.MQDurableName, s.NatsAddress, s.ClusterID)

	return sub
}

func (s *Service) stop(sub stan.Subscription) {
	sub.Close()

	_gZaplogSugar.Info("Gracefully finished.")
}

/******************************************************************************
 * Функции
 */

func messageProcessor(msg *stan.Msg) {
	svcMsg := &types.Message{}

	err := json.Unmarshal(msg.Data, svcMsg)
	if err != nil {
		_gZaplogSugar.Errorf("Cannot unmarshal. Error: %s. Msg: %s",
			err, spew.Sprintf("%#v", msg))

		return
	}

	_gZaplogSugar.Debug(spew.Sprintf("Msg: %#v", msg))

	dbMsg := convertServiceMsgToDatabaseMsg(svcMsg, msg.Timestamp)

	err = _gTelebot.SendAchtungMessage(dbMsg)
	if err != nil {
		_gZaplogSugar.Errorf("telebot cannot process message. Error: %s. Message: %s", err, spew.Sprintf("%#v", dbMsg))

		return
	}

	err = msg.Ack()
	if err != nil {
		_gZaplogSugar.Errorf("Cannot Ack message. Error: %s. Message: %s", err, spew.Sprintf("%#v", msg))
	}
}

func convertServiceMsgToDatabaseMsg(msg *types.Message, timestamp int64) *database.Message {
	return &database.Message{
		Subsystem: msg.Subsystem,
		Message:   msg.Message,
		CreatedAt: time.Unix(0, timestamp),
	}
}
