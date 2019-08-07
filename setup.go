package gogm

import (
	"errors"
	"github.com/cornelk/hashmap"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
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

//convert these into concurrent hashmap
var mappedTypes = &hashmap.HashMap{}
//relationship + label
var mappedRelations = &relationConfigs{}

func makeRelMapKey(start, edge, direction, rel string) string{
	return fmt.Sprintf("%s-%s-%v-%s", start, edge, direction, rel)
}

var isSetup = false

func Init(conf *Config, mapTypes ...interface{}) error {
	return setupInit(false, conf, mapTypes...)
}

func setupInit(isTest bool, conf *Config, mapTypes ...interface{}) error{
	if isSetup && !isTest{
		return errors.New("gogm has already been initialized")
	}
	if !isTest{
		if conf == nil{
			return errors.New("config can not be nil")
		}
	}

	log.Debug("mapping types")
	for _, t := range mapTypes{
		name := reflect.TypeOf(t).Elem().Name()
		dc, rels, err := getStructDecoratorConfig(t)
		if err != nil{
			return err
		}

		log.Debugf("mapped type '%s'", name)

		if len(rels) > 0{
			for k, v := range rels{
				mappedRelations.Set(k, v)
			}
		}

		log.Infof("mapped type %s", name)
		mappedTypes.Set(name, *dc)
	}

	if !isTest{
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
	}


	log.Debug("starting index verification step")
	if !isTest{
		var err error
		if conf.IndexStrategy == ASSERT_INDEX{
			log.Debug("chose ASSERT_INDEX strategy")
			log.Debug("dropping all known indexes")
			err = dropAllIndexesAndConstraints()
			if err != nil{
				return err
			}

			log.Debug("creating all mapped indexes")
			err = createAllIndexesAndConstraints(mappedTypes)
			if err != nil{
				return err
			}

			log.Debug("verifying all indexes")
			err = verifyAllIndexesAndConstraints(mappedTypes)
			if err != nil {
				return err
			}
		} else if conf.IndexStrategy == VALIDATE_INDEX{
			log.Debug("chose VALIDATE_INDEX strategy")
			log.Debug("verifying all indexes")
			err = verifyAllIndexesAndConstraints(mappedTypes)
			if err != nil {
				return err
			}
		} else if conf.IndexStrategy == IGNORE_INDEX{
			log.Debug("ignoring indexes")
		} else {
			return errors.New("unknown index strategy")
		}
	}


	log.Debug("setup complete")

	isSetup = true

	return nil
}


