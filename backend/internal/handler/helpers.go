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
	pp, err := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if err != nil || pp <= 0 {
		return 25
	}
	switch pp {
	case 12, 20, 24, 25, 48, 50, 96, 100:
		return pp
	default:
		if pp > 1000 {
			return 1000
		}
		return pp
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
