package engine

import (
	stderrors "errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/pipeline"
)

type engineBase struct {
	Config       config.Interface
	PipelineData *pipeline.Data
}

//Helper functions

func (e *engineBase) GenerateNextVersion(currentVersion string) (string, error) {
	v, nerr := semver.NewVersion(currentVersion)
	if nerr != nil {
		return "", nerr
	}

	switch bumpType := e.Config.GetString(config.PACKAGR_VERSION_BUMP_TYPE); bumpType {
	case "major":
		return fmt.Sprintf("%d.%d.%d", v.Major()+1, 0, 0), nil
	case "minor":
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor()+1, 0), nil
	case "patch":
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch()+1), nil
	default:
		return "", stderrors.New("Unknown version bump interval")
	}

}
