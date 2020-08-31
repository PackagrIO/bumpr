// +build python

package engine_test

import (
	"github.com/analogj/go-util/utils"
	"github.com/golang/mock/gomock"
	"github.com/packagrio/bumpr/pkg/config"
	"github.com/packagrio/bumpr/pkg/engine"
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

func TestEnginePython_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "python")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create("github", pipelineData)
	require.NoError(t, err)

	//test
	pythonEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_PYTHON, pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, pythonEngine)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EnginePythonTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EnginePythonTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EnginePythonTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnginePython_TestSuite(t *testing.T) {
	suite.Run(t, new(EnginePythonTestSuite))
}

func (suite *EnginePythonTestSuite) TestEnginePython_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	//suite.Config.EXPECT().GetBool("engine_disable_lint").Return(false)
	//suite.Config.EXPECT().GetBool("engine_disable_security_check").Return(false)

	pythonEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_PYTHON, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EnginePythonTestSuite) TestEnginePython_BumpVersion() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).Return("patch").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_METADATA_PATH).Return("VERSION").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "pip_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create("python", suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.BumpVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EnginePythonTestSuite) TestEnginePython_BumpVersion_WithMinimalRepo() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).Return("patch").MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_METADATA_PATH).Return("VERSION").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "minimal_pip_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	pythonEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_PYTHON, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.BumpVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "VERSION")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "tox.ini")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, "tests", "__init__.py")))
	require.True(suite.T(), utils.FileExists(path.Join(suite.PipelineData.GitLocalPath, ".gitignore")))
}

func (suite *EnginePythonTestSuite) TestEnginePython_BumpVersion_WithoutSetupPy() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "pip_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "python", "minimal_pip_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "setup.py"))

	pythonEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_PYTHON, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := pythonEngine.BumpVersion()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}