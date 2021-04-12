package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"rss-graber/pkg/api"
	"rss-graber/pkg/models"
	"rss-graber/pkg/rss"
	"rss-graber/pkg/store"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	_ "github.com/lib/pq"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	_host       = "localhost"
	_port       = 5432
	_user       = "postgres"
	_password   = "mysecretpassword"
	_dbname     = "rssitem"
	_portServer = "1300"
)

//Store Работа с БД
type Store interface {
	CreateDB(string) error
	CreateTables() error
	AddItems(item <-chan models.Item) //error
	AddResourse(string) (int, error)
	GetItems([]string, int, int) ([]models.Item, error)
	Close()
}

var (
	sugarLogger *zap.SugaredLogger
)

//InitLogger Конфигурирование логирования
func initLogger() {
	writerSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())
	sugarLogger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./pkg/log/Debug-info.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func init() {
	initLogger()
}

func main() {
	sugarLogger.Infof("Successs: %s", "Program is starting")

	GetedItem := models.NewRWItem()

	MyStore := deployNewDataBase()
	go MyStore.AddItems(GetedItem.GetedItem)
	defer MyStore.Close()
	defer GetedItem.CloseChan()

	serverFunc := api.DefRealization{Decoder: schema.NewDecoder(),
		DataHighway: GetedItem.GetedItem,
		Source:      rss.NewRss(),
		Store:       MyStore,
	}

	router := mux.NewRouter()

	var api = router.PathPrefix("/api").Subrouter()
	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	//Проверка работоспособности сервера
	api.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ServerRun"))
	})

	var apiV1 = api.PathPrefix("/v1").Subrouter()
	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	apiV1.HandleFunc("/AddUrl", serverFunc.AddUrlRss).Methods("POST")

	apiV1.HandleFunc("/GetItems", serverFunc.GetItems).Methods("GET")

	log.Fatal(http.ListenAndServe(":"+_portServer, router))
}

//Разворачивание пустой базы
func deployNewDataBase() Store {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", _host, _port, _user, _password)
	db, err := setConn(connString)
	if err != nil {
		CriticalError(err)
	}
	var tempStore Store = &store.DBStore{DB: db}

	//Создание новой базы данных
	err = tempStore.CreateDB(_dbname)
	if err != nil {
		CriticalError(err)
	}
	tempStore.Close()
	//Создание таблиц
	connString += fmt.Sprintf(" dbname=%s ", _dbname)
	db, err = setConn(connString)
	if err != nil {
		CriticalError(err)
	}
	tempStore = &store.DBStore{DB: db}
	err = tempStore.CreateTables()
	if err != nil {
		CriticalError(err)
	}
	return tempStore
}

//Инициализация нового соединения
func setConn(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		sugarLogger.Errorf("DB Connection: %s", connString)
		return nil, err
	}
	sugarLogger.Infof("DB Connection: %s", connString)
	return db, nil
}

//CriticalError Критическая ошибка, дальнейшее использование программы невозможно
func CriticalError(err error) {
	sugarLogger.Panicf("CriticalError %s", err)

	os.Exit(1)
}
