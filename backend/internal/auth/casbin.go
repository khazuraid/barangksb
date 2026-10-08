package auth

import (
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (p.act == r.act || p.act == "*")
`

const rbacPolicy = `
p, admin, /api/*, *
p, petugas, /api/auth/me, GET
p, petugas, /api/auth/password, POST
p, petugas, /api/dashboard, GET
p, petugas, /api/items, GET
p, petugas, /api/items/:id, GET
p, petugas, /api/items, POST
p, petugas, /api/items/template*, GET
p, petugas, /api/items/import, POST
p, petugas, /api/items/:id, PUT
p, petugas, /api/items/:id/maintenance*, *
p, petugas, /api/maintenance/*, GET
p, petugas, /api/categories, GET
p, petugas, /api/locations, GET
p, petugas, /api/movement/*, POST
p, petugas, /api/transactions, GET
p, petugas, /api/barcode/*, GET
p, petugas, /api/upload, POST
p, petugas, /api/export/*, GET
p, petugas, /api/report*, GET
p, petugas, /api/adjust/bulk, POST
p, petugas, /api/audit, GET
p, petugas, /api/settings/telegram, GET
p, petugas, /api/settings/telegram/bot, GET
p, petugas, /api/settings/telegram/subscribers, GET
p, petugas, /api/settings/telegram/logs, GET
p, petugas, /api/settings/telegram/commands, GET
`

func NewEnforcer() (*casbin.Enforcer, error) {
	m, err := model.NewModelFromString(rbacModel)
	if err != nil {
		return nil, fmt.Errorf("casbin model: %w", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("casbin enforcer: %w", err)
	}
	for _, line := range strings.Split(rbacPolicy, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) >= 4 && parts[0] == "p" {
			e.AddPolicy(parts[1], parts[2], parts[3])
		}
	}
	return e, nil
}