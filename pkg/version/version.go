// Copyright Contributors to the Open Cluster Management project

package version

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Masterminds/semver"
)

// Version is the semver version the operator is reconciling towards
var Version string

func init() {
	if value, exists := os.LookupEnv("OPERATOR_VERSION"); exists {
		Version = value
	} else {
		Version = "9.9.9"
	}
}

// Info contains versioning information.
type Info struct {
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() Info {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the settings in pkg/version/base.go
	return Info{
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// Older returns true if the version is older than this package's running version
func Older(v string) (bool, error) {
	operatorVersion, err := semver.NewVersion(Version)
	if err != nil {
		return false, fmt.Errorf("could not parse operator version into semver version: %s", Version)
	}

	currentVersion, err := semver.NewVersion(v)
	if err != nil {
		return false, fmt.Errorf("could not parse supplied version into semver version: %s", v)
	}

	return operatorVersion.GreaterThan(currentVersion), nil
}
