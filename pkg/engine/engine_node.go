package engine

import (
	"encoding/json"
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"io/ioutil"
	"os/exec"
	"path"
)

type engineNode struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.NodeMetadata
	NextMetadata    *metadata.NodeMetadata
}

func (g *engineNode) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.NodeMetadata)
	g.NextMetadata = new(metadata.NodeMetadata)

	return nil
}

func (g *engineNode) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineNode) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineNode) ValidateTools() error {

	if _, kerr := exec.LookPath("node"); kerr != nil {
		return errors.EngineValidateToolError("node binary is missing")
	}

	return nil
}

func (g *engineNode) BumpVersion() error {

	// bump up the package version
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.writeNextMetadata(g.PipelineData.GitLocalPath); nerr != nil {
		return nerr
	}

	return nil
}

//private Helpers

func (g *engineNode) retrieveCurrentMetadata(gitLocalPath string) error {
	//read package.json file.
	packageContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, "package.json"))
	if rerr != nil {
		return rerr
	}

	if uerr := json.Unmarshal(packageContent, g.CurrentMetadata); uerr != nil {
		return uerr
	}

	return nil
}

func (g *engineNode) populateNextMetadata() error {

	nextVersion, err := g.GenerateNextVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.NextMetadata.Name = g.CurrentMetadata.Name
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineNode) writeNextMetadata(gitLocalPath string) error {
	// The version will be bumped up via the npm version command.
	// --no-git-tag-version ensures that we dont create a git commit (which npm will do by default).
	versionCmd := fmt.Sprintf("npm --no-git-tag-version version %s",
		g.NextMetadata.Version,
	)
	if verr := utils.BashCmdExec(versionCmd, g.PipelineData.GitLocalPath, nil, ""); verr != nil {
		return errors.EngineTestRunnerError("npm version bump failed")
	}
	return nil
}
