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

const PACKAGR_ENGINE_TYPE_CHEF = "chef"
const PACKAGR_ENGINE_TYPE_GENERIC = "generic"
const PACKAGR_ENGINE_TYPE_GOLANG = "golang"
const PACKAGR_ENGINE_TYPE_NODE = "node"
const PACKAGR_ENGINE_TYPE_PYTHON = "python"
const PACKAGR_ENGINE_TYPE_RUBY = "ruby"
