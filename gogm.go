// Copyright (c) 2021 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gogm

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/cornelk/hashmap"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
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
	return NewContext(context.Background(), config, pkStrategy, mapTypes...)
}

func NewContext(ctx context.Context, config *Config, pkStrategy *PrimaryKeyStrategy, mapTypes ...interface{}) (*Gogm, error) {
	if config == nil {
		return nil, errors.New("config can not be nil")
	}

	if pkStrategy == nil {
		return nil, errors.New("pk strategy can not be nil")
	}

	if len(mapTypes) == 0 {
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
		pkStrategy:      pkStrategy,
	}

	err := g.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init gogm instance, %w", err)
	}

	return g, nil
}

func (g *Gogm) init(ctx context.Context) error {
	err := g.validate()
	if err != nil {
		return err
	}

	err = g.parseOgmTypes()
	if err != nil {
		return err
	}

	g.logger.Debug("establishing neo connection")

	err = g.initDriver(ctx)
	if err != nil {
		return err
	}

	g.logger.Debug("initializing indices")
	return g.initIndex(ctx)
}

func (g *Gogm) validate() error {
	err := g.config.validate()
	if err != nil {
		return fmt.Errorf("config failed validation, %w", err)
	}

	g.logger = g.config.Logger

	if g.config.TargetDbs == nil || len(g.config.TargetDbs) == 0 {
		g.config.TargetDbs = []string{"neo4j"}
	}

	if g.pkStrategy == nil {
		// setting to the default pk strategy
		g.pkStrategy = DefaultPrimaryKeyStrategy
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

func (g *Gogm) initDriver(ctx context.Context) error {
	var certPool *x509.CertPool
	isEncrypted := strings.Contains(g.config.Protocol, "+s")

	if isEncrypted {
		if g.config.UseSystemCertPool {
			var err error
			certPool, err = x509.SystemCertPool()
			if err != nil {
				return fmt.Errorf("failed to get system cert pool")
			}
		} else {
			certPool = x509.NewCertPool()
		}

		if g.config.CAFileLocation != "" {
			bytes, err := ioutil.ReadFile(g.config.CAFileLocation)
			if err != nil {
				return fmt.Errorf("failed to open ca file, %w", err)
			}

			certPool.AppendCertsFromPEM(bytes)
		}
	}

	neoConfig := func(neoConf *neo4j.Config) {
		if g.config.EnableDriverLogs {
			neoConf.Log = wrapLogger(g.logger)
		}

		neoConf.MaxConnectionPoolSize = g.config.PoolSize

		if isEncrypted {
			neoConf.RootCAs = certPool
		}
	}

	doneChan := make(chan error, 1)

	_, hasDeadline := ctx.Deadline()

	go g.initDriverRoutine(neoConfig, doneChan)

	if hasDeadline {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		select {
		case err := <-doneChan:
			if err != nil {
				return fmt.Errorf("failed to init driver, %w", err)
			}
			return nil
		case <-ctx.Done():
			return errors.New("timed out initializing driver")
		}
	} else {
		err := <-doneChan
		if err != nil {
			return fmt.Errorf("failed to init driver, %w", err)
		}
		return nil
	}
}

func (g *Gogm) initDriverRoutine(neoConfig func(neoConf *neo4j.Config), doneChan chan error) {
	driver, err := neo4j.NewDriver(g.config.ConnectionString(), neo4j.BasicAuth(g.config.Username, g.config.Password, g.config.Realm), neoConfig)
	if err != nil {
		doneChan <- fmt.Errorf("failed to create driver, %w", err)
		return
	}

	err = driver.VerifyConnectivity()
	if err != nil {
		doneChan <- fmt.Errorf("failed to verify connectivity, %w", err)
		return
	}

	// set driver
	g.driver = driver

	// get neoversion
	sess := driver.NewSession(neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
		//	DatabaseName: "neo4j",
	})

	res, err := sess.Run("return 1", nil)
	if err != nil {
		doneChan <- err
		return
	} else if err = res.Err(); err != nil {
		doneChan <- err
		return
	}

	sum, err := res.Consume()
	if err != nil {
		doneChan <- err
		return
	}

	version := strings.Split(strings.Replace(strings.ToLower(sum.Server().Version()), "neo4j/", "", -1), ".")
	g.neoVersion, err = strconv.ParseFloat(version[0], 64)
	if err != nil {
		doneChan <- err
		return
	}
	doneChan <- nil
}

func (g *Gogm) initIndex(ctx context.Context) error {
	switch g.config.IndexStrategy {
	case ASSERT_INDEX:
		g.logger.Debug("chose ASSERT_INDEX strategy")
		g.logger.Debug("dropping all known indexes")
		err := dropAllIndexesAndConstraints(ctx, g)
		if err != nil {
			return err
		}

		g.logger.Debug("creating all mapped indexes")
		err = createAllIndexesAndConstraints(ctx, g, g.mappedTypes)
		if err != nil {
			return err
		}

		g.logger.Debug("verifying all indexes")
		err = verifyAllIndexesAndConstraints(ctx, g, g.mappedTypes)
		if err != nil {
			return err
		}
		return nil
	case VALIDATE_INDEX:
		g.logger.Debug("chose VALIDATE_INDEX strategy")
		g.logger.Debug("verifying all indexes")
		err := verifyAllIndexesAndConstraints(ctx, g, g.mappedTypes)
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
