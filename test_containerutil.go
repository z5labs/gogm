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
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

type neo4jContainerVersion int

const (
	neo3 neo4jContainerVersion = iota
	neo4
)

type neo4jContainer struct {
	testcontainers.Container
	CertDir string
}

const neo4jContainerUsername = "neo4j"
const neo4jContainerPassword = "changeme"

func setupNeo(ctx context.Context, version neo4jContainerVersion) (*neo4jContainer, error) {
	var err error

	req := testcontainers.ContainerRequest{
		Image:        "neo4j:3.5-enterprise",
		ExposedPorts: []string{"7474/tcp", "6477/tcp", "7687/tcp"},
		Env: map[string]string{
			"NEO4J_ACCEPT_LICENSE_AGREEMENT": "yes",
			"NEO4J_AUTH":                     neo4jContainerUsername + "/" + neo4jContainerPassword,
		},
	}

	// defaults to empty string if using neo3 (no certs)
	certDir := ""

	if version == neo4 {
		certDir, err = createTempNeoKeypair()
		if err != nil {
			return nil, err
		}

		req.Image = "neo4j:4.4-enterprise"

		req.Env["NEO4J_dbms_memory_heap_initial__size"] = "512m"
		req.Env["NEO4J_dbms_memory_heap_max__size"] = "2G"
		req.Env["NEO4J_dbms_default__listen__address"] = "0.0.0.0"

		req.Env["NEO4J_dbms_allow__upgrade"] = "true"
		req.Env["NEO4J_dbms_default__database"] = "neo4j"

		req.Env["NEO4J_dbms_ssl_policy_bolt_enabled"] = "true"
		req.Env["NEO4J_dbms_ssl_policy_bolt_base__directory"] = "/ssl/client_policy"
		req.Env["NEO4J_dbms_connector_bolt_tls__level"] = "OPTIONAL"

		// mount certificate directory
		req.Mounts = testcontainers.ContainerMounts{
			testcontainers.BindMount(certDir, "/ssl/client_policy"),
		}
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &neo4jContainer{Container: container, CertDir: certDir}, nil
}

func (n *neo4jContainer) Teminate(ctx context.Context) error {
	var keyPairCleanupErr error
	if n.CertDir != "" {
		keyPairCleanupErr = cleanupTempNeoKeypair(n.CertDir)
	}

	containerCleanupErr := n.Container.Terminate(ctx)

	if keyPairCleanupErr != nil && containerCleanupErr != nil {
		return fmt.Errorf("Container termination (%v) and keypair cleanup (%v) failed", containerCleanupErr, keyPairCleanupErr)
	} else if keyPairCleanupErr != nil {
		return keyPairCleanupErr
	} else if containerCleanupErr != nil {
		return containerCleanupErr
	}

	return nil
}
