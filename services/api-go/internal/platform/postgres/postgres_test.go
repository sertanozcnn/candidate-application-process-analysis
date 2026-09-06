package postgres

import (
	"context"
	"testing"
)

func TestConnectRejectsInvalidDatabaseURL(t *testing.T) {
	_, err := Connect(context.Background(), "not-a-url")
	if err == nil {
		t.Fatal("Connect() error = nil, want error")
	}
}
