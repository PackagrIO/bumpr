// +build golang

package engine_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/golang/mock/gomock"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/bumpr/pkg/engine"
	"github.com/packagrio/go-common/metadata"
	"github.com/packagrio/go-common/pipeline"
	"github.com/packagrio/go-common/scm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"path"
	//"path/filepath"
	"github.com/packagrio/bumpr/pkg/config/mock"
	"github.com/packagrio/go-common/scm/mock"
	"os"
	"testing"
)

func TestEngineGolang_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "golang")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData)
	require.NoError(t, err)

	//test
	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, golangEngine)
	require.Equal(t, "exit 0", testConfig.GetString("engine_cmd_security_check"), "should load engine defaults")
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineGolangTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineGolangTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineGolangTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineGolang_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineGolangTestSuite))
}

func (suite *EngineGolangTestSuite) TestEngineGolang_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_SCM).Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)

	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_BumpVersion() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).Return("patch").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_SCM).Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_METADATA_PATH).Return("pkg/version/version.go").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.BumpVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "1.0.1", golangEngine.GetNextMetadata().(metadata.GolangMetadata).Version)

}

func (suite *EngineGolangTestSuite) TestEngineGolang_BumpVersion_WithMinimalRepo() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).Return("patch").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_SCM).Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_METADATA_PATH).Return("pkg/version/version.go").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "minimal_golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.BumpVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "1.0.1", golangEngine.GetNextMetadata().(metadata.GolangMetadata).Version)
}

func (suite *EngineGolangTestSuite) TestEngineGolang_BumpVersion_WithoutVersion() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_SCM).Return("github").MinTimes(1)
	suite.Config.EXPECT().GetString("scm_repo_full_name").Return("AnalogJ/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString("engine_golang_package_path").Return("github.com/analogj/golang_analogj_test").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_METADATA_PATH).Return("pkg/version/version.go").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "golang_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "golang", "minimal_golang_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "pkg", "version", "version.go"))

	golangEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_GOLANG, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := golangEngine.BumpVersion()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}
