package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bsergik/telebot-nats/internal/database"
	"github.com/bsergik/telebot-nats/internal/logger"
	"github.com/bsergik/telebot-nats/internal/service"
	"github.com/bsergik/telebot-nats/internal/telebot"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

/******************************************************************************
 * Types
 */

type environments struct {
	DryRun        bool   `long:"dry-run" env:"BOT_DRY_RUN" description:"do not send data to telegram bot"`
	TelegramToken string `long:"telegram-token" env:"BOT_TELEGRAM_TOKEN" description:"set telegram-token of a bot" required:"true"`
	MasterUserID  int    `long:"master-user-id" env:"BOT_MASTER_USER_ID" description:"set master user id" required:"true"`
	NatsAddress   string `long:"nats-address" env:"BOT_NATS_ADDRESS" description:"address of NATS server" required:"true" default:"nats://localhost:4222"`

	DBHost     string `long:"db-host" env:"BOT_DB_HOST" description:"set host of Postgres DB" required:"true"`
	DBPort     uint   `long:"db-port" env:"BOT_DB_PORT" description:"set port of Postgres DB" default:"5432"`
	DBName     string `long:"db-name" env:"BOT_DB_NAME" description:"set database name for Postgres DB" required:"true"`
	DBUser     string `long:"db-user" env:"BOT_DB_USER" description:"set user of Postgres DB" required:"true"`
	DBPassword string `long:"db-password" env:"BOT_DB_PASSWORD" description:"set password of Postgres DB" required:"true"`
	DBInit     bool   `long:"init-db" env:"BOT_INIT_DB" description:"create required tables"`

	ClusterID        string `long:"cluster-id" env:"BOT_CLUSTERID" description:"cluster ID on STAN" required:"true"`
	ClientID         string `long:"client-id" env:"BOT_CLIENTID" description:"client ID on STAN" required:"true"`
	MQSubjectName    string `long:"subject" env:"BOT_SUBJECT" description:"set subject to subscribe to on STAN" default:"telebot.v1.errors"`
	MQQueueGroupName string `long:"queue-group" env:"BOT_QUEUE_GROUP" description:"set queue group name" default:"telebot.v1.qgroup"`
	MQDurableName    string `long:"durable-name" env:"BOT_DURABLE_NAME" description:"set durable name" default:"telebot.v1.qgroup-durable"`
}

/******************************************************************************
 * Constants
 */

const (
	dbReconnectTimeout = 12
)

/******************************************************************************
 * Global variables
 */

var (
	buildVersion  string
	_gZaplogSugar *zap.SugaredLogger
)

/******************************************************************************
 * Functions
 */

func main() {
	logger, _ := logger.NewLogger()
	_gZaplogSugar = logger.Sugar()

	defer func() { _ = logger.Sync() }()

	/** */

	envs := getEnvironments()
	printDebug(envs)

	db := dbConnect(envs)
	defer db.Disconnect()

	finishCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	goRoutinesNumber := 2

	tb := startTelebot(logger, envs, wg, finishCh, db)
	startService(logger, envs, wg, finishCh, db, tb)
	startSignalHandler(finishCh, goRoutinesNumber)

	wg.Wait()
}

func getEnvironments() *environments {
	var err error

	e := new(environments)
	parser := flags.NewParser(e, flags.Default)

	if _, err = parser.Parse(); err != nil {
		_gZaplogSugar.Fatalf("getEnvs: %s", err)
	}

	return e
}

func printDebug(envs *environments) {
	_gZaplogSugar.Debugf("Version: %s", buildVersion)
	_gZaplogSugar.Debugf("telegramToken len: %d", len(envs.TelegramToken))
	_gZaplogSugar.Debugf("dbHost: %s", envs.DBHost)
	_gZaplogSugar.Debugf("DBPort: %d", envs.DBPort)
	_gZaplogSugar.Debugf("dbName: %s", envs.DBName)
	_gZaplogSugar.Debugf("dbUser: %s", envs.DBUser)
	_gZaplogSugar.Debugf("dbPassword len: %d", len(envs.DBPassword))
}

func dbConnect(envs *environments) database.IStorage {
	for {
		db, err := database.Connect(envs.DBHost, envs.DBName, envs.DBUser, envs.DBPassword, envs.DBPort)
		if err != nil {
			time.Sleep(dbReconnectTimeout * time.Second)

			continue
		}

		if envs.DBInit {
			err := db.InitDB()
			if err != nil {
				_gZaplogSugar.Errorf("Cannot init DB. Error: %s", err)
			}
		}

		return db
	}
}

func startService(logger *zap.Logger, e *environments, wg *sync.WaitGroup, finish <-chan struct{}, storage database.IStorage, tb *telebot.Telebot) {
	wg.Add(1)

	go (&service.Service{
		Zaplog:  logger,
		Storage: storage,
		Telebot: tb,

		NatsAddress: e.NatsAddress,
		ClusterID:   e.ClusterID,
		ClientID:    e.ClientID,

		MQSubjectName:    e.MQSubjectName,
		MQQueueGroupName: e.MQQueueGroupName,
		MQDurableName:    e.MQDurableName,
	}).Start(wg, finish)
}

func startTelebot(logger *zap.Logger, e *environments, wg *sync.WaitGroup, finish <-chan struct{}, storage database.IStorage) *telebot.Telebot {
	wg.Add(1)

	tb := &telebot.Telebot{
		Zaplog:        logger,
		TelegramToken: e.TelegramToken,
		MasterUserID:  e.MasterUserID,
		Storage:       storage,
	}

	go tb.Start(wg, finish)

	return tb
}

func startSignalHandler(finish chan<- struct{}, goRoutinesNumber int) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		_gZaplogSugar.Debugf("Signal %d received. Gracefully stopping ...", sig)

		for i := 0; i < goRoutinesNumber; i++ {
			finish <- struct{}{}
		}

		close(sigs)
	}
}
