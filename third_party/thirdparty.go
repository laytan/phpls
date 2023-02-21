// Package thirdparty exports the embedded filesystem of phpstorm-stubs.
package thirdparty

import "embed"

//go:embed phpstorm-stubs
var Stubs embed.FS
