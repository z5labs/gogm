package gogm

import (
	"errors"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/sirupsen/logrus"
	"reflect"
)

var externalLog *logrus.Entry

var log = getLogger()

func getLogger() *logrus.Entry{
	if externalLog == nil{
		//create default logger
		toReturn := logrus.New()

		return toReturn.WithField("source", "gogm")
	}

	return externalLog
}

func SetLogger(logger *logrus.Entry) error {
	if logger == nil{
		return errors.New("logger can not be nil")
	}
	externalLog = logger
	return nil
}

type Config struct{
	Host string
	Port int

	Username string
	Password string

	PoolSize int

	IndexStrategy IndexStrategy
}

type IndexStrategy int

const (
	ASSERT_INDEX IndexStrategy = iota
	VALIDATE_INDEX
	IGNORE_INDEX
)

var mappedTypes = map[string]structDecoratorConfig{}

var isSetup = false

func Init(conf *Config, mapTypes ...interface{}) error{
	if isSetup{
		return errors.New("gogm has already been initialized")
	}

	if conf == nil{
		return errors.New("config can not be nil")
	}

	log.Debug("mapping types")
	for _, t := range mapTypes{
		name := reflect.TypeOf(t).Name()
		dc, err := getStructDecoratorConfig(t)
		if err != nil{
			return err
		}

		log.Debugf("mapped type '%s'", name)

		mappedTypes[name] = *dc
	}

	log.Debug("opening connection to neo4j")
	err := dsl.Init(&dsl.ConnectionConfig{
		PoolSize: conf.PoolSize,
		Port: conf.Port,
		Host: conf.Host,
		Password: conf.Password,
		Username: conf.Username,
	})
	if err != nil{
		return err
	}

	log.Debug("starting index verification step")

	if conf.IndexStrategy == ASSERT_INDEX{
		log.Debug("chose ASSERT_INDEX strategy")
		log.Debug("dropping all known indexes")
		err = dropAllIndexes()
		if err != nil{
			return err
		}

		log.Debug("creating all mapped indexes")
		err = createAllIndexes()
		if err != nil{
			return err
		}

		log.Debug("verifying all indexes")
		err = verifyAllIndexes()
		if err != nil {
			return err
		}
	} else if conf.IndexStrategy == VALIDATE_INDEX{
		log.Debug("chose VALIDATE_INDEX strategy")
		log.Debug("verifying all indexes")
		err = verifyAllIndexes()
		if err != nil {
			return err
		}
	} else if conf.IndexStrategy == IGNORE_INDEX{
		log.Debug("ignoring indexes")
	} else {
		return errors.New("unknown index strategy")
	}

	log.Debug("setup complete")

	isSetup = true

	return nil
}


