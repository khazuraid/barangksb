package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func fmtSprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

func configPerPage(c *gin.Context) int {
	pp, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	switch pp {
	case 20, 50, 100:
		return pp
	default:
		return 25
	}
}

type paginatedResp struct {
	Data    any `json:"data"`
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Pages   int `json:"pages"`
}

func nullable(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
