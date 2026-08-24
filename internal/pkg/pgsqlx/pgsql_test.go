package pgsqlx

import (
	"nurture/internal/config"
	"testing"
)

func TestInitPgsqlDisabled(t *testing.T) {
	old := config.Conf.DB
	t.Cleanup(func() {
		config.Conf.DB = old
	})
	config.Conf.DB = config.DB{Enable: false}

	if pool := InitPgsql(); pool != nil {
		t.Fatal("InitPgsql() returned pool when db is disabled")
	}
}
