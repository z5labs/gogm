// Copyright (c) 2022 MindStand Technologies, Inc
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
	"crypto/tls"
	"errors"
	"fmt"
	"time"
)

const (
	defaultRetryWait = time.Second * 10
)

// Config defines parameters for creating a GoGM object
type Config struct {
	// Host is the neo4j host
	Host string `yaml:"host" json:"host" mapstructure:"host"`
	// Port is the neo4j port
	Port int `yaml:"port" json:"port" mapstructure:"port"`

	// deprecated in favor of Protocol
	// IsCluster specifies whether GoGM is connecting to a casual cluster or not, will determine whether to use bolt or neo4j protocols
	IsCluster bool `yaml:"is_cluster" json:"is_cluster" mapstructure:"is_cluster"`

	// Protocol specifies which protocol gogm will connect to neo4j with
	// The options are neo4j, neo4j+s, neo4j+ssc, bolt, bolt+s and bolt+ssc
	Protocol string `json:"protocol" yaml:"protocol" mapstructure:"protocol"`

	// Username is the GoGM username
	Username string `yaml:"username" json:"username" mapstructure:"username"`

	// Password is the GoGM password
	Password string `yaml:"password" json:"password" mapstructure:"password"`

	// PoolSize is the size of the connection pool for GoGM
	PoolSize int `yaml:"pool_size" json:"pool_size" mapstructure:"pool_size"`

	// DefaultTransactionTimeout defines the default time a transaction will wait before timing out
	DefaultTransactionTimeout time.Duration `json:"default_transaction_timeout" yaml:"default_transaction_timeout" mapstructure:"default_transaction_timeout"`

	// Realm defines the realm passed into neo4j
	Realm string `yaml:"realm" json:"realm" mapstructure:"realm"`

	// deprecated: in favor of TLSConfig
	//these security configurations will be ignored if the protocol does not contain +s
	UseSystemCertPool bool `yaml:"use_system_cert_pool" mapstructure:"use_system_cert_pool"`

	// deprecated: in favor of TLSConfig
	// CAFileLocation defines the location of the CA file for authenticating with tls
	CAFileLocation string `yaml:"ca_file_location" mapstructure:"ca_file_location"`

	// TLSConfig defines the configuration for connecting to a neo4j cluster over tls
	TLSConfig *tls.Config `yaml:"tls_config" mapstructure:"tls_config"`

	// Index Strategy defines the index strategy for GoGM
	// Options for index strategy are:
	// IGNORE_INDEX - which does no index/constraint operations
	// VALIDATE_INDEX - which validates whether the indexes/constraints exist
	// ASSERT_INDEX - which deletes existing indexes/constraints for the given nodes then creates them
	IndexStrategy IndexStrategy `yaml:"index_strategy" json:"index_strategy" mapstructure:"index_strategy"`

	// TargetDbs tells gogm which databases to expect and is also what index operations use to know which dbs to execute against
	TargetDbs []string `yaml:"target_dbs" json:"target_dbs" mapstructure:"target_dbs"`

	// Logger specifies log interfaces that gogm will use to log
	Logger Logger `yaml:"-" json:"-" mapstructure:"-"`

	// LogLevel defines the log level that the logger will use
	// if logger is not nil log level will be ignored
	LogLevel string `json:"log_level" yaml:"log_level" mapstructure:"log_level"`

	// EnableDriverLogs tells the gogm whether to log logs coming out of the neo4j go driver
	EnableDriverLogs bool `json:"enable_driver_logs" yaml:"enable_driver_logs" mapstructure:"enable_driver_logs"`

	// EnableLogParams tells gogm whether to log params going into queries when on debug/trace log level
	// WARNING THIS IS A SECURITY RISK -- ONLY ENABLE THIS FOR DEBUG
	EnableLogParams bool `json:"enable_log_properties" yaml:"enable_log_properties" mapstructure:"enable_log_properties"`

	// OpentracingEnabled tells gogm whether to use open tracing
	OpentracingEnabled bool `json:"opentracing_enabled" yaml:"opentracing_enabled" mapstructure:"opentracing_enabled"`

	// LoadStrategy tells gogm how to generate load queries
	// The options are:
	// PATH_LOAD_STRATEGY - this generates queries based on path `match p=...`. The queries are less verbose than schema but generally slower
	// SCHEMA_LOAD_STRATEGY - this generates queries based on the gogm schema. The queries are a lot more verbose but will generally execute faster
	LoadStrategy LoadStrategy `json:"load_strategy" yaml:"load_strategy" mapstructure:"load_strategy"`
}

// validate checks whether config object params are valid
func (c *Config) validate() error {
	if c.Logger == nil {
		c.Logger = GetDefaultLogger()
	}

	if c.DefaultTransactionTimeout <= 0 {
		// default is 1 second
		c.DefaultTransactionTimeout = defaultRetryWait
	}

	if c.Host == "" {
		return errors.New("hostname not defined")
	}

	if c.Port <= 0 {
		return errors.New("port either not specified or invalid")
	}

	if c.TargetDbs == nil || len(c.TargetDbs) == 0 {
		c.TargetDbs = []string{"neo4j"}
	}

	if err := c.IndexStrategy.validate(); err != nil {
		return err
	}

	if err := c.LoadStrategy.validate(); err != nil {
		return err
	}

	return nil
}

// ConnectionString builds the neo4j connection string
func (c *Config) ConnectionString() string {
	var protocol string

	if c.Protocol != "" {
		protocol = c.Protocol
	} else {
		if c.IsCluster {
			protocol = "neo4j"
		} else {
			protocol = "bolt"
		}
	}

	// In case of special characters in password string
	//password := url.QueryEscape(c.Password)
	return fmt.Sprintf("%s://%s:%v", protocol, c.Host, c.Port)
}

// IndexStrategy defines the different index approaches
type IndexStrategy int

const (
	// ASSERT_INDEX ensures that all indices are set and sets them if they are not there
	ASSERT_INDEX IndexStrategy = 0
	// VALIDATE_INDEX ensures that all indices are set
	VALIDATE_INDEX IndexStrategy = 1
	// IGNORE_INDEX skips the index step of setup
	IGNORE_INDEX IndexStrategy = 2
)

func (is IndexStrategy) validate() error {
	switch is {
	case ASSERT_INDEX, VALIDATE_INDEX, IGNORE_INDEX:
		return nil
	default:
		return fmt.Errorf("invalid index strategy %d", is)
	}
}
