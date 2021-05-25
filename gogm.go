package gogm

import (
	"errors"
	"fmt"
	"github.com/cornelk/hashmap"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"reflect"
	"strconv"
	"strings"
)

var globalGogm = &Gogm{isNoOp: true, logger: GetDefaultLogger()}

// SetGlobalGogm sets the global instance of gogm
func SetGlobalGogm(gogm *Gogm) {
	globalGogm = gogm
}

// G returns the global instance of gogm
func G() *Gogm {
	return globalGogm
}

type Gogm struct {
	config          *Config
	pkStrategy      *PrimaryKeyStrategy
	logger          Logger
	neoVersion      float64
	mappedTypes     *hashmap.HashMap
	driver          neo4j.Driver
	mappedRelations *relationConfigs
	ogmTypes        []interface{}
	// isNoOp specifies whether this instance of gogm can do anything
	// is only used for the default global gogm
	isNoOp bool
}

func New(config *Config, pkStrategy *PrimaryKeyStrategy, mapTypes ...interface{}) (*Gogm, error) {
	if config == nil {
		return nil, errors.New("config can not be nil")
	}

	if pkStrategy == nil {
		return nil, errors.New("pk strategy can not be nil")
	}

	if mapTypes == nil || len(mapTypes) == 0 {
		return nil, errors.New("no types to map")
	}

	g := &Gogm{
		config:          config,
		logger:          config.Logger,
		neoVersion:      0,
		mappedTypes:     &hashmap.HashMap{},
		driver:          nil,
		mappedRelations: &relationConfigs{},
		ogmTypes:        mapTypes,
	}

	err := g.init()
	if err != nil {
		return nil, fmt.Errorf("failed to init gogm instance, %w", err)
	}

	return g, nil
}

func (g *Gogm) init() error {
	err := g.validate()
	if err != nil {
		return err
	}

	err = g.parseOgmTypes()
	if err != nil {
		return err
	}

	g.logger.Debug("establishing neo connection")
	err = g.initDriver()
	if err != nil {
		return err
	}

	g.logger.Debug("initializing indices")
	return g.initIndex()
}

func (g *Gogm) validate() error {
	err := g.config.validate()
	if err != nil {
		return fmt.Errorf("config failed validation, %w", err)
	}

	if g.config.TargetDbs == nil || len(g.config.TargetDbs) == 0 {
		g.config.TargetDbs = []string{"neo4j"}
	}

	if g.pkStrategy == nil {
		return errors.New("pkStrategy can not be nil")
	}

	err = g.pkStrategy.validate()
	if err != nil {
		return fmt.Errorf("pk strategy failed validation, %w", err)
	}

	return nil
}

func (g *Gogm) parseOgmTypes() error {
	g.logger.Debug("mapping types")
	for _, t := range g.ogmTypes {
		name := reflect.TypeOf(t).Elem().Name()
		dc, err := getStructDecoratorConfig(g, t, g.mappedRelations)
		if err != nil {
			return err
		}

		g.logger.Debugf("mapped type %s", name)
		g.mappedTypes.Set(name, *dc)
	}

	// validate relationships
	g.logger.Debug("validating edges")
	err := g.mappedRelations.Validate()
	if err != nil {
		g.logger.Debugf("failed to validate edges, %v", err)
		return fmt.Errorf("failed to validate edges, %w", err)
	}

	return nil
}

func (g *Gogm) initDriver() error {
	// todo tls support
	neoConfig := func(neoConf *neo4j.Config) {
		if g.config.EnableDriverLogs {
			neoConf.Log = wrapLogger(g.logger)
		}

		neoConf.MaxConnectionPoolSize = g.config.PoolSize
	}

	driver, err := neo4j.NewDriver(g.config.ConnectionString(), neo4j.BasicAuth(g.config.Username, g.config.Password, g.config.Realm), neoConfig)
	if err != nil {
		return fmt.Errorf("failed to create driver, %w", err)
	}

	err = driver.VerifyConnectivity()
	if err != nil {
		return fmt.Errorf("failed to verify connectivity, %w", err)
	}

	// set driver
	g.driver = driver

	// get neoversion
	sess := driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
	//	DatabaseName: "neo4j",
	})

	res, err := sess.Run("return 1", nil)
	if err != nil {
		return err
	} else if err = res.Err(); err != nil {
		return err
	}

	sum, err := res.Consume()
	if err != nil {
		return err
	}

	version := strings.Split(strings.Replace(strings.ToLower(sum.Server().Version()), "neo4j/", "", -1), ".")
	g.neoVersion, err = strconv.ParseFloat(version[0], 64)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gogm) initIndex() error {
	switch g.config.IndexStrategy {
	case ASSERT_INDEX:
		g.logger.Debug("chose ASSERT_INDEX strategy")
		g.logger.Debug("dropping all known indexes")
		err := dropAllIndexesAndConstraints(g)
		if err != nil {
			return err
		}

		g.logger.Debug("creating all mapped indexes")
		err = createAllIndexesAndConstraints(g, g.mappedTypes)
		if err != nil {
			return err
		}

		g.logger.Debug("verifying all indexes")
		err = verifyAllIndexesAndConstraints(g, g.mappedTypes)
		if err != nil {
			return err
		}
		return nil
	case VALIDATE_INDEX:
		g.logger.Debug("chose VALIDATE_INDEX strategy")
		g.logger.Debug("verifying all indexes")
		err := verifyAllIndexesAndConstraints(g, g.mappedTypes)
		if err != nil {
			return err
		}
		return nil
	case IGNORE_INDEX:
		g.logger.Debug("ignoring indices")
		return nil
	default:
		g.logger.Debugf("unknown index strategy, %v", g.config.IndexStrategy)
		return fmt.Errorf("unknown index strategy, %v", g.config.IndexStrategy)
	}
}

func (g *Gogm) Copy() *Gogm {
	return &Gogm{
		config:          g.config,
		logger:          g.logger,
		neoVersion:      g.neoVersion,
		mappedTypes:     g.mappedTypes,
		driver:          g.driver,
		mappedRelations: g.mappedRelations,
		ogmTypes:        g.ogmTypes,
	}
}

func (g *Gogm) Close() error {
	if g.driver == nil {
		return errors.New("unable to close nil driver")
	}

	return g.driver.Close()
}

func (g *Gogm) NewSession(conf SessionConfig) (ISession, error) {
	if g.isNoOp {
		return nil, errors.New("gogm instance is no op. Please set global logger with SetGlobalLogger() or create a new gogm instance")
	}

	return newSessionWithConfig(g, conf)
}

func (g *Gogm) NewSessionV2(conf SessionConfig) (SessionV2, error) {
	if g.isNoOp {
		return nil, errors.New("gogm instance is no op. Please set global logger with SetGlobalLogger() or create a new gogm instance")
	}

	return newSessionWithConfigV2(g, conf)
}
