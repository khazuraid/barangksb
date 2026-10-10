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
	val := strings.TrimSpace(c.DefaultQuery("per_page", "25"))
	if val == "all" || val == "-1" || val == "semua" || c.Query("all") == "true" {
		return 10000
	}
	pp, err := strconv.Atoi(val)
	if err != nil || pp <= 0 {
		return 25
	}
	if pp > 10000 {
		return 10000
	}
	return pp
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
