package auth

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// RBAC model: sub (role), obj (path), act (method)
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
p, petugas, /api/items, GET
p, petugas, /api/items/:id, GET
p, petugas, /api/items, POST
p, petugas, /api/items/:id, PUT
p, petugas, /api/categories, GET
p, petugas, /api/locations, GET
p, petugas, /api/movement/*, POST
p, petugas, /api/transactions, GET
p, petugas, /api/barcode/*, GET
p, petugas, /api/upload, POST
p, petugas, /api/export/*, GET
p, petugas, /api/report*, GET
p, petugas, /api/dashboard, GET
p, petugas, /api/adjust/bulk, POST
p, petugas, /api/drivesync, GET
p, petugas, /api/auth/me, GET
p, petugas, /api/password, POST
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
	// load policies from string
	lines := splitLines(rbacPolicy)
	for _, line := range lines {
		if line == "" || line[0] == '#' {
			continue
		}
		parts := splitComma(line)
		if len(parts) >= 4 && parts[0] == "p" {
			e.AddPolicy(parts[1], parts[2], parts[3])
		}
	}
	return e, nil
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[0] != ' ' {
				out = append(out, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i, c := range s {
		if c == ',' {
			part := s[start:i]
			// trim spaces
			for len(part) > 0 && part[0] == ' ' {
				part = part[1:]
			}
			for len(part) > 0 && part[len(part)-1] == ' ' {
				part = part[:len(part)-1]
			}
			out = append(out, part)
			start = i + 1
		}
	}
	if start < len(s) {
		part := s[start:]
		for len(part) > 0 && part[0] == ' ' {
			part = part[1:]
		}
		for len(part) > 0 && part[len(part)-1] == ' ' {
			part = part[:len(part)-1]
		}
		out = append(out, part)
	}
	return out
}
