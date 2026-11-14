package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
)

type fakeConnector struct{}

func (fakeConnector) Connect(context.Context) (driver.Conn, error) { return fakeConn{}, nil }
func (fakeConnector) Driver() driver.Driver                        { return nil }

type fakeConn struct{}

func (fakeConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (fakeConn) Close() error                        { return nil }
func (fakeConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func TestConfigurePoolBoundsOpenAndIdleConnections(t *testing.T) {
	sqlDB := sql.OpenDB(fakeConnector{})
	defer sqlDB.Close()
	configurePool(sqlDB)

	if got := sqlDB.Stats().MaxOpenConnections; got != maxOpenConns {
		t.Fatalf("expected max open connections %d, got %d", maxOpenConns, got)
	}

	// Check out more connections than the idle limit, then return them all.
	ctx := context.Background()
	conns := make([]*sql.Conn, 0, maxIdleConns+5)
	for i := 0; i < maxIdleConns+5; i++ {
		conn, err := sqlDB.Conn(ctx)
		if err != nil {
			t.Fatalf("checkout %d failed: %v", i, err)
		}
		conns = append(conns, conn)
	}
	for _, conn := range conns {
		_ = conn.Close()
	}

	stats := sqlDB.Stats()
	if stats.Idle != maxIdleConns {
		t.Fatalf("expected %d idle connections retained, got %d", maxIdleConns, stats.Idle)
	}
	if stats.MaxIdleClosed != 5 {
		t.Fatalf("expected 5 connections closed above the idle limit, got %d", stats.MaxIdleClosed)
	}
}
