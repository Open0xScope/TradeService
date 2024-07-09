package config

import (
	"sync"

	"github.com/Open0xScope/CommuneXService/utils/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// one database one instance
type PostgresqlConfig struct {
	Host       string
	Port       int64
	Account    string
	Password   string
	DBName     string
	SchemaName string
}

type RedisConfig struct {
	Name         string
	Password     string
	Host         string
	DB           int64
	MinIdleConns int64
}

// struct decode must has tag
type Config struct {
	PostgresqlConfig PostgresqlConfig
	RedisConf        RedisConfig
}

var (
	configMutex = sync.RWMutex{}
	config      Config

	configViper     *viper.Viper
	configFlyChange []chan bool
)

func RegistConfChange(c chan bool) {
	configFlyChange = append(configFlyChange, c)
}

func notifyConfChange() {
	for i := 0; i < len(configFlyChange); i++ {
		configFlyChange[i] <- true
	}
}

func watchConfig(c *viper.Viper) error {
	c.WatchConfig()
	cfn := func(e fsnotify.Event) {
		logger.Logrus.WithFields(logrus.Fields{"change": e.String()}).Info("config change and reload it")
		reloadConfig(c)
		notifyConfChange()
	}

	c.OnConfigChange(cfn)
	return nil
}

func LoadConf(configFilePath string) error {
	config = Config{}
	configMutex.Lock()
	defer configMutex.Unlock()

	configViper = viper.New()
	configViper.SetConfigName("config")
	configViper.AddConfigPath(configFilePath) //endwith "/"
	configViper.SetConfigType("yaml")

	if err := configViper.ReadInConfig(); err != nil {
		return err
	}
	if err := configViper.Unmarshal(&config); err != nil {
		return err
	}

	logger.Logrus.WithFields(logrus.Fields{"Config": config}).Info("Load config success")

	if err := watchConfig(configViper); err != nil {
		return err
	}
	return nil
}

func reloadConfig(c *viper.Viper) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if err := c.ReadInConfig(); err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err.Error()}).Error("config ReLoad failed")
	}

	if err := configViper.Unmarshal(&config); err != nil {
		logger.Logrus.WithFields(logrus.Fields{"ErrMsg": err.Error()}).Error("unmarshal config failed")
	}

	logger.Logrus.WithFields(logrus.Fields{"config": config}).Info("Config ReLoad Success")
}

func GetPostgresqlConfig() PostgresqlConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return config.PostgresqlConfig
}

func GetRedisConfig() RedisConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return config.RedisConf
}
