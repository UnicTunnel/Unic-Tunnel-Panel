// Package web bundles the panel's HTML templates and static assets into the binary.
package web

import "embed"

//go:embed templates/*.html static/*
var FS embed.FS
