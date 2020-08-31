// +build chef

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
	mock_scm "github.com/packagrio/go-common/scm/mock"
	"os"
	"testing"
)

func TestEngineChef_Create(t *testing.T) {
	//setup
	testConfig, err := config.Create()
	require.NoError(t, err)

	testConfig.Set(config.PACKAGR_SCM, "github")
	testConfig.Set(config.PACKAGR_PACKAGE_TYPE, "chef")
	pipelineData := new(pipeline.Data)
	githubScm, err := scm.Create(engine.PACKAGR_ENGINE_TYPE_CHEF, pipelineData, testConfig, nil)
	require.NoError(t, err)

	//test
	chefEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_CHEF, pipelineData, testConfig, githubScm)

	//assert
	require.NoError(t, err)
	require.NotNil(t, chefEngine)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EngineChefTestSuite struct {
	suite.Suite
	MockCtrl     *gomock.Controller
	Scm          *mock_scm.MockInterface
	Config       *mock_config.MockInterface
	PipelineData *pipeline.Data
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *EngineChefTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	suite.PipelineData = new(pipeline.Data)

	suite.Config = mock_config.NewMockInterface(suite.MockCtrl)
	suite.Scm = mock_scm.NewMockInterface(suite.MockCtrl)

}

func (suite *EngineChefTestSuite) TearDownTest() {
	suite.MockCtrl.Finish()
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEngineChef_TestSuite(t *testing.T) {
	suite.Run(t, new(EngineChefTestSuite))
}

func (suite *EngineChefTestSuite) TestEngineChef_ValidateTools() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	chefEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_CHEF, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.ValidateTools()

	//assert
	require.NoError(suite.T(), berr)
}

func (suite *EngineChefTestSuite) TestEngineChef_BumpVersion() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).Return("patch").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_CHEF, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.BumpVersion()
	require.NoError(suite.T(), berr)

	//assert
	require.Equal(suite.T(), "")
}

func (suite *EngineChefTestSuite) TestEngineChef_BumpVersion_WithMinimalCookbook() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)
	suite.Config.EXPECT().GetString(config.PACKAGR_VERSION_BUMP_TYPE).Return("patch").MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "minimal_cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)

	chefEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_CHEF, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.BumpVersion()
	require.NoError(suite.T(), berr)

}

func (suite *EngineChefTestSuite) TestEngineChef_BumpVersion_WithoutMetadata() {
	//setup
	suite.Config.EXPECT().SetDefault(gomock.Any(), gomock.Any()).MinTimes(1)

	//copy cookbook fixture into a temp directory.
	parentPath, err := ioutil.TempDir("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(parentPath)
	suite.PipelineData.GitParentPath = parentPath
	suite.PipelineData.GitLocalPath = path.Join(parentPath, "cookbook_analogj_test")
	cerr := utils.CopyDir(path.Join("testdata", "chef", "minimal_cookbook_analogj_test"), suite.PipelineData.GitLocalPath)
	require.NoError(suite.T(), cerr)
	os.Remove(path.Join(suite.PipelineData.GitLocalPath, "metadata.rb"))

	chefEngine, err := engine.Create(engine.PACKAGR_ENGINE_TYPE_CHEF, suite.PipelineData, suite.Config, suite.Scm)
	require.NoError(suite.T(), err)

	//test
	berr := chefEngine.BumpVersion()

	//assert
	require.Error(suite.T(), berr, "should return an error")
}
