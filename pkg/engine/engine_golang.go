package engine

import (
	"bytes"
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/go-common/errors"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type engineGolang struct {
	engineBase

	Scm             scm.Interface //Interface
	CurrentMetadata *metadata.GolangMetadata
	NextMetadata    *metadata.GolangMetadata
}

func (g *engineGolang) Init(pipelineData *pipeline.Data, configData config.Interface, sourceScm scm.Interface) error {
	g.Scm = sourceScm
	g.Config = configData
	g.PipelineData = pipelineData
	g.CurrentMetadata = new(metadata.GolangMetadata)
	g.NextMetadata = new(metadata.GolangMetadata)

	//set command defaults (can be overridden by repo/system configuration)
	g.Config.SetDefault(config.PACKAGR_VERSION_METADATA_PATH, "pkg/version/version.go")
	var scmDomain string
	if g.Config.GetString(config.PACKAGR_SCM) == "bitbucket" {
		scmDomain = "bitbucket.org"
	} else {
		scmDomain = "github.com"
	}

	g.Config.SetDefault("engine_golang_package_path", fmt.Sprintf("%s/%s", scmDomain, strings.ToLower(g.Config.GetString("scm_repo_full_name"))))

	//TODO: figure out why setting the GOPATH workspace is causing the tools to timeout.
	// golang recommends that your in-development packages are in the GOPATH and glide requires it to do glide install.
	// the problem with this is that for somereason gometalinter (and the underlying linting tools) take alot longer
	// to run, and hit the default deadline limit ( --deadline=30s).
	// we can have multiple workspaces in the gopath by separating them with colon (:), but this timeout is nasty if not required.
	//TODO: g.GoPath root will not be deleted (its the parent of GitParentPath), figure out if we can do this automatically.
	g.PipelineData.GolangGoPath = g.PipelineData.GitParentPath
	os.Setenv("GOPATH", fmt.Sprintf("%s:%s", os.Getenv("GOPATH"), g.PipelineData.GolangGoPath))

	// A proper gopath has a bin and src directory.
	goPathBin := path.Join(g.PipelineData.GitParentPath, "bin")
	goPathSrc := path.Join(g.PipelineData.GitParentPath, "src")
	os.MkdirAll(goPathBin, 0666)
	os.MkdirAll(goPathSrc, 0666)

	//  the gopath bin directory should aslo be added to Path
	os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), goPathBin))

	packagePathPrefix := path.Dir(g.Config.GetString("engine_golang_package_path")) //strip out the repo name.
	// customize the git parent path for Golang Engine
	g.PipelineData.GitParentPath = path.Join(g.PipelineData.GitParentPath, "src", packagePathPrefix)
	os.MkdirAll(g.PipelineData.GitParentPath, 0666)

	return nil
}

func (g *engineGolang) GetCurrentMetadata() interface{} {
	return g.CurrentMetadata
}
func (g *engineGolang) GetNextMetadata() interface{} {
	return g.NextMetadata
}

func (g *engineGolang) ValidateTools() error {
	if _, kerr := exec.LookPath("go"); kerr != nil {
		return errors.EngineValidateToolError("go binary is missing")
	}

	return nil
}

func (g *engineGolang) BumpVersion() error {
	//validate that the chef metadata.rb file exists

	if !utils.FileExists(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH))) {
		return errors.EngineBuildPackageInvalid(fmt.Sprintf("%s file is required to process Go library", g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	}

	// bump up the go package version
	if merr := g.retrieveCurrentMetadata(g.PipelineData.GitLocalPath); merr != nil {
		return merr
	}

	if perr := g.populateNextMetadata(); perr != nil {
		return perr
	}

	if nerr := g.SetVersion(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)), g.NextMetadata.Version); nerr != nil {
		return nerr
	}

	return nil
}

func (g *engineGolang) SetVersion(versionMetadataPath string, nextVersion string) error {
	return g.writeNextMetadata(versionMetadataPath, nextVersion)
}

//private Helpers

func (g *engineGolang) retrieveCurrentMetadata(gitLocalPath string) error {

	versionContent, rerr := ioutil.ReadFile(path.Join(g.PipelineData.GitLocalPath, g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
	if rerr != nil {
		return rerr
	}

	//Oh.My.God.

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(versionContent), 0)
	if err != nil {
		return err
	}

	version, verr := g.parseGoVersion(f.Decls)
	if verr != nil {
		return verr
	}

	g.CurrentMetadata.Version = version
	return nil
}

func (g *engineGolang) populateNextMetadata() error {

	nextVersion, err := g.GenerateNextVersion(g.CurrentMetadata.Version)
	if err != nil {
		return err
	}

	g.NextMetadata.Version = nextVersion
	g.PipelineData.ReleaseVersion = g.NextMetadata.Version
	return nil
}

func (g *engineGolang) writeNextMetadata(gitLocalMetadataPath string, nextVersion string) error {
	versionPath := gitLocalMetadataPath
	versionContent, rerr := ioutil.ReadFile(versionPath)
	if rerr != nil {
		return rerr
	}

	//Oh.My.God.

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(versionContent), parser.ParseComments)
	if err != nil {
		return err
	}

	decls, serr := g.setGoVersion(f.Decls, nextVersion)
	if serr != nil {
		return serr
	}
	f.Decls = decls

	//write the version file again.
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return err
	}

	return ioutil.WriteFile(versionPath, buf.Bytes(), 0644)
}

func (g *engineGolang) parseGoVersion(list []ast.Decl) (string, error) {
	//find version declaration (uppercase or lowercase)
	for _, decl := range list {
		gen := decl.(*ast.GenDecl)
		if gen.Tok == token.CONST || gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valSpec := spec.(*ast.ValueSpec)
				if strings.ToLower(valSpec.Names[0].Name) == "version" {
					//found the version variable.
					return strings.Trim(valSpec.Values[0].(*ast.BasicLit).Value, "\"'"), nil
				}
			}
		}
	}
	return "", errors.EngineBuildPackageFailed(fmt.Sprintf("Could not retrieve the version from %s", g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
}

func (g *engineGolang) setGoVersion(list []ast.Decl, version string) ([]ast.Decl, error) {
	//find version declaration (uppercase or lowercase)
	for _, decl := range list {
		gen := decl.(*ast.GenDecl)
		if gen.Tok == token.CONST || gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valSpec := spec.(*ast.ValueSpec)
				if strings.ToLower(valSpec.Names[0].Name) == "version" {
					//found the version variable.
					valSpec.Values[0].(*ast.BasicLit).Value = fmt.Sprintf(`"%s"`, version)
					return list, nil
				}
			}
		}
	}
	return nil, errors.EngineBuildPackageFailed(fmt.Sprintf("Could not set the version in %s", g.Config.GetString(config.PACKAGR_VERSION_METADATA_PATH)))
}
