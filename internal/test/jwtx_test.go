package test

import (
	"nurture/internal/pkg/jwtx"
	"testing"
)

func TestJwtx(t *testing.T) {
	token1, err := jwtx.GenTestToken("1", jwtx.COMMON_USER)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token1)

	token2, err := jwtx.GenTestToken("2", jwtx.COMMON_USER)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token2)
}
