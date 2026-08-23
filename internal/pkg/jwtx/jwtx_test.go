package jwtx

import (
	"testing"
)

func TestJwtx(t *testing.T) {
	token1, err := GenTestToken("1", COMMON_USER)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token1)

	token2, err := GenTestToken("2", COMMON_USER)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token2)
}
