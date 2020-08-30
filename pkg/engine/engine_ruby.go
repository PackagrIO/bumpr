package engine

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
)

type rubyGemspec struct {
	Name    string `json:"name"`
	Version struct {
		Version string `json:"name"`
	} `json:"version"`
}

type engineRuby struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.RubyMetadata
	NextMetadata    *metadata.RubyMetadata
	GemspecPath     string
}

func (g *engineRuby) Init(pipelineData *pipeline.Data, config config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = config
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.RubyMetadata)
	g.NextMetadata = new(metadata.RubyMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	return nil
}

func (g *engineRuby) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineRuby) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineRuby) ValidateTools() error {
	if _, kerr := exec.LookPath("ruby"); kerr != nil {
		return errors.EngineValidateToolError("ruby binary is missing")
	}

	return nil
}

func (g *engineRuby) BumpVersion() error {

	// bump up the version here.
	// since there's no standardized way to bump up the version in the *.gemspec file, we're going to assume that the version
	// is specified in a version file in the lib/<gem_name>/ directory, similar to how the bundler gem does it.
	// ie. bundle gem <gem_name> will create a file: my_gem/lib/my_gem/version.rb with the following contents
	// module MyGem
	//   VERSION = "0.1.0"
	// end
	//
	// Jeweler and Hoe both do something similar.
	// http://yehudakatz.com/2010/04/02/using-gemspecs-as-intended/
	// http://timelessrepo.com/making-ruby-gems
	// http://guides.rubygems.org/make-your-own-gem/
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
func (g *engineRuby) retrieveCurrentMetadata(gitLocalPath string) error {
	//read Gemspec file.
	gemspecFiles, gerr := filepath.Glob(path.Join(gitLocalPath, "/*.gemspec"))
	if gerr != nil {
		return errors.EngineBuildPackageInvalid("*.gemspec file is required to process Ruby gem")
	} else if len(gemspecFiles) == 0 {
		return errors.EngineBuildPackageInvalid("*.gemspec file is required to process Ruby gem")
	}

	g.GemspecPath = gemspecFiles[0]

	gemspecJsonFile, _ := ioutil.TempFile("", "gemspec.json")
	defer os.Remove(gemspecJsonFile.Name())

	//generate a JSON-style YAML file containing the Gemspec data. (still not straight valid JSON).
	//
	gemspecJsonCmd := fmt.Sprintf("ruby -e \"require('yaml'); File.write('%s', YAML::to_json(Gem::Specification::load('%s')))\"",
		gemspecJsonFile.Name(),
		g.GemspecPath,
	)
	if cerr := utils.BashCmdExec(gemspecJsonCmd, "", nil, ""); cerr != nil {
		return errors.EngineBuildPackageFailed(fmt.Sprintf("Command (%s) failed. Check log for more details.", gemspecJsonCmd))
	}

	//Load gemspec JSON file and parse it.
	gemspecJsonContent, rerr := ioutil.ReadFile(gemspecJsonFile.Name())
	if rerr != nil {
		return rerr
	}

	gemspecObj := new(rubyGemspec)
	if uerr := yaml.Unmarshal(gemspecJsonContent, gemspecObj); uerr != nil {
		fmt.Println(string(gemspecJsonContent))
		return uerr
	}

	g.CurrentMetadata.Name = gemspecObj.Name
	g.CurrentMetadata.Version = gemspecObj.Version.Version

	//ensure that there is a lib/GEMNAME/version.rb file.
	versionrbPath := path.Join("lib", gemspecObj.Name, "version.rb")
	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, versionrbPath)) {
		return errors.EngineBuildPackageInvalid(
			fmt.Sprintf("version.rb file (%s) is required to process Ruby gem", versionrbPath))
	}
	return nil
}

func (g *engineRuby) populateNextMetadata() error {

	nextVersion, err := g.GenerateNextVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.NextMetadata.Name = g.CurrentMetadata.Name
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineRuby) writeNextMetadata(gitLocalPath string) error {

	versionrbPath := path.Join(g.PipelineData.GitLocalPath, "lib", g.CurrentMetadata.Name, "version.rb")
	versionrbContent, rerr := ioutil.ReadFile(versionrbPath)
	if rerr != nil {
		return rerr
	}
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	updatedContent := re.ReplaceAllLiteralString(string(versionrbContent), g.NextMetadata.Version)
	return ioutil.WriteFile(versionrbPath, []byte(updatedContent), 0644)
}
