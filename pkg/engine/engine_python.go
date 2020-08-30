package engine

import (
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"
)

type enginePython struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.PythonMetadata
	NextMetadata    *metadata.PythonMetadata
}

func (g *enginePython) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.PythonMetadata)
	g.NextMetadata = new(metadata.PythonMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "VERSION")
	return nil
}

func (g *enginePython) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *enginePython) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *enginePython) ValidateTools() error {
	if _, berr := exec.LookPath("python"); berr != nil {
		return errors.EngineValidateToolError("python binary is missing")
	}

	return nil
}

func (g *enginePython) BumpVersion() error {
	//validate that the python setup.py file exists
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, "setup.py")) {
		return errors.EngineBuildPackageInvalid("setup.py file is required to process Python package")
	}

	// check for/create required VERSION file
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH))) {
		ioutil.WriteFile(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)),
			[]byte("0.0.0"),
			0644,
		)
	}

	// bump up the version here.
	// since there's no standardized way to bump up the version in the setup.py file, we're going to assume that the version
	// is specified in plain text VERSION file in the root of the source repository. This can be configured via version_metadata_path
	// this is option #4 in the python packaging guide:
	// https://packaging.python.org/en/latest/single_source_version/#single-sourcing-the-version
	//
	// additional packaging structures, like those listed below, may also be supported in the future.
	// http://stackoverflow.com/a/7071358/1157633

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

func (g *enginePython) retrieveCurrentMetadata(gitLocalPath string) error {
	//read metadata.json file.
	versionContent, rerr := ioutil.ReadFile(path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}
	g.CurrentMetadata.Version = strings.TrimSpace(string(versionContent))
	return nil
}

func (g *enginePython) populateNextMetadata() error {

	nextVersion, err := g.GenerateNextVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *enginePython) writeNextMetadata(gitLocalPath string) error {
	return ioutil.WriteFile(path.Join(gitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)), []byte(g.NextMetadata.Version), 0644)
}
