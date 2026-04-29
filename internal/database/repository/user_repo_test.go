package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestUpdateStatus_ReturnsErrorWhenUserDoesNotExist(t *testing.T) {
	db, cleanup := openStubDB(t, stubExecResult{rowsAffected: 0})
	defer cleanup()

	repo := &UserRepository{db: db}

	err := repo.UpdateStatus(42, "blocked")
	if err == nil {
		t.Fatal("expected error when updating nonexistent user, got nil")
	}
	if !strings.Contains(err.Error(), "user 42 not found") {
		t.Fatalf("expected not found error, got %q", err.Error())
	}
}

type stubExecResult struct {
	rowsAffected int64
	err          error
}

type stubDriver struct {
	result stubExecResult
}

func (d *stubDriver) Open(name string) (driver.Conn, error) {
	return &stubConn{result: d.result}, nil
}

type stubConn struct {
	result stubExecResult
}

func (c *stubConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("prepare not implemented")
}

func (c *stubConn) Close() error {
	return nil
}

func (c *stubConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not implemented")
}

func (c *stubConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if query != "UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?" {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 args, got %d", len(args))
	}
	if got := args[0].Value; got != "blocked" {
		return nil, fmt.Errorf("expected status arg blocked, got %v", got)
	}
	if got := args[1].Value; got != int64(42) {
		return nil, fmt.Errorf("expected user id arg 42, got %v", got)
	}
	if c.result.err != nil {
		return nil, c.result.err
	}
	return driver.RowsAffected(c.result.rowsAffected), nil
}

func (c *stubConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("query not implemented")
}

var stubDriverCounter int

func openStubDB(t *testing.T, result stubExecResult) (*sql.DB, func()) {
	t.Helper()

	driverName := fmt.Sprintf("stub-user-repo-%d", stubDriverCounter)
	stubDriverCounter++
	sql.Register(driverName, &stubDriver{result: result})

	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close stub db: %v", err)
		}
	}

	return db, cleanup
}

var _ driver.ExecerContext = (*stubConn)(nil)
var _ driver.QueryerContext = (*stubConn)(nil)
var _ driver.Conn = (*stubConn)(nil)
var _ driver.Driver = (*stubDriver)(nil)
var _ driver.Rows = (*stubRows)(nil)

type stubRows struct{}

func (r *stubRows) Columns() []string { return nil }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next([]driver.Value) error {
	return io.EOF
}
