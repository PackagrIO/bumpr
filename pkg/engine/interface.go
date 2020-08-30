package engine

import (
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
)

// Create mock using:
// mockgen -source=pkg/engine/interface.go -destination=pkg/engine/mock/mock_engine.go
type Interface interface {
	Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error

	// Validate that required executables are available for the following build/test/package/etc steps
	ValidateTools() error

	BumpVersion() error

	GetCurrentMetadata() interface{}
	GetNextMetadata() interface{}
}
