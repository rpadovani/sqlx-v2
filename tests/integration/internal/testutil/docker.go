// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tc_toxiproxy "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgOnce      sync.Once
	pgDSN       string
	pgContainer *postgres.PostgresContainer

	myOnce      sync.Once
	myDSN       string
	myContainer *mysql.MySQLContainer

	toOnce      sync.Once
	toContainer *tc_toxiproxy.Container
	toProxy     *toxiproxy.Proxy
)

// HasDocker check if the local system has a running Docker daemon.
// If it fails, tests relying on Postgres/MySQL will be skipped.
func HasDocker() bool {
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GlobalPostgres returns a shared Postgres DSN for the entire test process.
func GlobalPostgres(t *testing.T) string {
	if t != nil {
		t.Helper()
		if !HasDocker() {
			t.Skip("Docker is not available; skipping Postgres integration tests")
		}
	} else if !HasDocker() {
		return "skip"
	}

	pgOnce.Do(func() {
		ctx := context.Background()
		c, err := postgres.Run(ctx,
			"postgres:15-alpine",
			postgres.WithDatabase("sqlxtest"),
			postgres.WithUsername("postgres"),
			postgres.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second),
			),
		)
		if err != nil {
			log.Fatalf("failed to start postgres container: %v", err)
		}
		pgContainer = c
		dsn, err := c.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			log.Fatalf("failed to get postgres connection string: %v", err)
		}
		pgDSN = dsn
	})

	return pgDSN
}

// GlobalToxiproxyPostgres returns a DSN for a Postgres instance proxied through Toxiproxy.
func GlobalToxiproxyPostgres(t *testing.T) (string, *toxiproxy.Proxy) {
	if t != nil {
		t.Helper()
		if !HasDocker() {
			t.Skip("Docker is not available; skipping Toxiproxy tests")
		}
	} else if !HasDocker() {
		return "skip", nil
	}

	// Ensure Postgres is running
	_ = GlobalPostgres(t)

	toOnce.Do(func() {
		ctx := context.Background()
		// Get the internal IP of the Postgres container
		pgIP, err := pgContainer.ContainerIP(ctx)
		if err != nil {
			log.Fatalf("failed to get postgres container IP: %v", err)
		}

		c, err := tc_toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.11.0",
			tc_toxiproxy.WithProxy("postgres", pgIP+":5432"),
		)
		if err != nil {
			log.Fatalf("failed to start toxiproxy container: %v", err)
		}
		toContainer = c

		// Connect to Toxiproxy control API
		toEndpoint, err := c.URI(ctx)
		if err != nil {
			log.Fatalf("failed to get toxiproxy URI: %v", err)
		}
		client := toxiproxy.NewClient(toEndpoint)

		// Get the proxy object (already created by WithProxy)
		p, err := client.Proxy("postgres")
		if err != nil {
			log.Fatalf("failed to get toxiproxy proxy: %v", err)
		}
		toProxy = p
	})

	// Get the port Toxiproxy is listening on inside its container
	_, portStr, err := net.SplitHostPort(toProxy.Listen)
	if err != nil {
		log.Fatalf("failed to split toxiproxy listen address %q: %v", toProxy.Listen, err)
	}
	var listenPort int
	if _, err := fmt.Sscanf(portStr, "%d", &listenPort); err != nil {
		log.Fatalf("failed to parse toxiproxy listen port %q: %v", portStr, err)
	}

	host, port, err := toContainer.ProxiedEndpoint(listenPort)
	if err != nil {
		log.Fatalf("failed to get toxiproxy host endpoint for port %d: %v", listenPort, err)
	}

	// Build DSN: postgres://postgres:postgres@<host>:<port>/sqlxtest?sslmode=disable
	dsn := "postgres://postgres:postgres@" + host + ":" + port + "/sqlxtest?sslmode=disable"
	return dsn, toProxy
}

// GlobalMySQL returns a shared MySQL DSN for the entire test process.
func GlobalMySQL(t *testing.T) string {
	if t != nil {
		t.Helper()
		if !HasDocker() {
			t.Skip("Docker is not available; skipping MySQL integration tests")
		}
	} else if !HasDocker() {
		return "skip"
	}

	myOnce.Do(func() {
		ctx := context.Background()
		c, err := mysql.Run(ctx,
			"mysql:8",
			mysql.WithDatabase("sqlxtest"),
			mysql.WithUsername("root"),
			mysql.WithPassword("root"),
		)
		if err != nil {
			log.Fatalf("failed to start mysql container: %v", err)
		}
		myContainer = c
		dsn, err := c.ConnectionString(ctx, "multiStatements=true", "parseTime=true")
		if err != nil {
			log.Fatalf("failed to get mysql connection string: %v", err)
		}
		myDSN = dsn
	})

	return myDSN
}

// CleanupContainers terminates any globally started containers.
func CleanupContainers() {
	ctx := context.Background()
	if pgContainer != nil {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}
	if myContainer != nil {
		if err := myContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate mysql container: %v", err)
		}
	}
	if toContainer != nil {
		if err := toContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate toxiproxy container: %v", err)
		}
	}
}
