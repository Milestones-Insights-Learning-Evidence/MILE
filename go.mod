module github.com/MOAJ-Mentor-of-all-Jacks/MOAJ

// Minimum language/API level. `go mod tidy` canonicalizes this directive to a full
// x.y.z version, and no `toolchain` line is declared, so this is a floor and not a
// build pin: CI selects the toolchain with `go-version: 1.26.x` in
// .github/workflows/ci-license-gate.yml so builds pick up the latest patched 1.26
// release instead of inheriting standard-library advisories fixed after 1.26.0.
go 1.26.0

require github.com/prometheus/client_golang v1.25.0

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.0 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
