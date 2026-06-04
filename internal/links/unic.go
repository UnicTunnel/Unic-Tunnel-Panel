// Package links builds and parses unic:// connection links.
// Format: unic://<base64url(json)>  where json = { v, name, host, port, user, password }.
// Canonical spec: docs/unic-link-spec.md. The App and Panel MUST agree on Version.
package links

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const Version = 1

type Payload struct {
	V        int    `json:"v"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func Build(p Payload) (string, error) {
	if p.V == 0 {
		p.V = Version
	}
	j, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return "unic://" + base64.RawURLEncoding.EncodeToString(j), nil
}

func Parse(link string) (*Payload, error) {
	if !strings.HasPrefix(link, "unic://") {
		return nil, fmt.Errorf("not a unic:// link")
	}
	raw := strings.TrimRight(strings.TrimPrefix(link, "unic://"), "=")
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("base64: %w", err)
	}
	var p Payload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	if p.V != Version {
		return nil, fmt.Errorf("unsupported link version %d (this build understands v=%d)", p.V, Version)
	}
	return &p, nil
}
