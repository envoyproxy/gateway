package buildflags

import "runtime"

var (
	// Do not edit, these are set at build time by goreleaser.yaml
	version string
	os      string
	arch    string
)

func GetVersion() string {
	if version == "" {
		return "0.0.0"
	}

	return version
}

func GetOS() string {
	if os == "" {
		return runtime.GOOS
	}
	return os
}

func GetArch() string {
	if arch == "" {
		return runtime.GOARCH
	}
	return arch
}
