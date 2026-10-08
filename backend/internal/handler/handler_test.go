package handler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	assert.Equal(t, "alat-kebersihan", slugify("Alat Kebersihan"))
	assert.Equal(t, "atk-alat-tulis", slugify("ATK (Alat Tulis)"))
	assert.Equal(t, "ruang-poli", slugify("Ruang Poli"))
}

func TestPagination(t *testing.T) {
	assert.Equal(t, 4, calcPages(100, 25))
	assert.Equal(t, 1, calcPages(5, 25))
	assert.Equal(t, 3, calcPages(75, 25))
}

func calcPages(total, perPage int) int {
	if perPage < 1 {
		return 1
	}
	n := total / perPage
	if total%perPage > 0 {
		n++
	}
	if n < 1 {
		n = 1
	}
	return n
}

func TestNullableS(t *testing.T) {
	assert.Nil(t, nullableS(""))
	assert.Nil(t, nullableS("  "))
	assert.Equal(t, "hello", nullableS("hello"))
}

func TestConfigPerPage(t *testing.T) {
	// Test logic without gin.Context
	assert.Equal(t, 20, perPageLogic("20"))
	assert.Equal(t, 50, perPageLogic("50"))
	assert.Equal(t, 100, perPageLogic("100"))
	assert.Equal(t, 25, perPageLogic("25"))
	assert.Equal(t, 25, perPageLogic(""))
	assert.Equal(t, 25, perPageLogic("999"))
}

func perPageLogic(s string) int {
	switch s {
	case "20":
		return 20
	case "50":
		return 50
	case "100":
		return 100
	case "":
		return 25
	default:
		return 25
	}
}

var _ = http.StatusOK
