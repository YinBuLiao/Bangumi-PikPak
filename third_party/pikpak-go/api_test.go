package pikpakgo_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	pikpakgo "github.com/kanghengliu/pikpak-go"

	"github.com/stretchr/testify/suite"
)

var (
	username = os.Getenv("PIKPAK_USERNAME")
	password = os.Getenv("PIKPAK_PASSWORD")

	// getAllFileFieldId: Make sure your field contains 100+ files. leave this env empty to skip get all files test.
	getAllFileFieldId = os.Getenv("PIKPAK_ALL_FILE_FIELD_ID")
)

type TestPikpakSuite struct {
	suite.Suite
	client *pikpakgo.PikPakClient
}

func TestPikpakAPI(t *testing.T) {
	suite.Run(t, new(TestPikpakSuite))
}

func (suite *TestPikpakSuite) SetupTest() {
	client, err := pikpakgo.NewPikPakClient(username, password)
	suite.NoError(err)
	err = client.Login()
	suite.NoError(err)
	suite.client = client
}

func (suite *TestPikpakSuite) TestListFile() {
	files, err := suite.client.FileList(100, "", "")
	suite.NoError(err)
	suite.NotEmpty(files)
}

func (suite *TestPikpakSuite) TestGetDownloadUrl() {
	files, err := suite.client.FileList(100, "", "")
	suite.NoError(err)
	suite.NotNil(files)
	suite.NotEmpty(files.Files)
	for _, f := range files.Files {
		println(fmt.Sprintf("%s %s", f.Kind, f.ID))
	}
	url, err := suite.client.GetDownloadUrl("VNWHVrXMz4En2yoLs_x-Uf_Ko1")
	suite.NoError(err)
	println(url)
}

func (suite *TestPikpakSuite) TestOfflineDownload() {
	task, err := suite.client.OfflineDownload("test", "magnet:?xt=urn:btih:bce204d9d53d7c843856b0b17c1d5dc1478d1cd5&tr=http%3a%2f%2ft.nyaatracker.com%2fannounce&tr=http%3a%2f%2ftracker.kamigami.org%3a2710%2fannounce&tr=http%3a%2f%2fshare.camoe.cn%3a8080%2fannounce&tr=http%3a%2f%2fopentracker.acgnx.se%2fannounce&tr=http%3a%2f%2fanidex.moe%3a6969%2fannounce&tr=http%3a%2f%2ft.acg.rip%3a6699%2fannounce&tr=https%3a%2f%2ftr.bangumi.moe%3a9696%2fannounce&tr=udp%3a%2f%2ftr.bangumi.moe%3a6969%2fannounce&tr=http%3a%2f%2fopen.acgtracker.com%3a1096%2fannounce&tr=udp%3a%2f%2ftracker.opentrackr.org%3a1337%2fannounce", "")
	suite.NoError(err)
	println(task.Task.ID)
	suite.NotEmpty(task.Task.ID)
	finishedTask, err := suite.client.WaitForOfflineDownloadComplete(task.Task.ID, time.Minute*1, nil)
	suite.NoError(err)
	println(finishedTask)
	uri, err := suite.client.GetDownloadUrl(finishedTask.FileID)
	suite.NoError(err)
	println(uri)
	uri2, err := suite.client.GetDownloadUrl("VNUHa9_xcQJe5gcxb7tZOpefo1")
	suite.NoError(err)
	println(uri2)
	files, err := suite.client.FileList(100, "", "")
	suite.NoError(err)
	suite.NotNil(files)
}

func (suite *TestPikpakSuite) TestOfflineList() {
	tasks, err := suite.client.OfflineList(100, "")
	suite.NoError(err)
	for _, f := range tasks.Tasks {
		println(fmt.Sprintf("%s %s", f.Kind, f.ID))
	}
}

func (suite *TestPikpakSuite) TestEmptyTrash() {
	err := suite.client.EmptyTrash()
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestTaskRetry() {
	tasks, err := suite.client.OfflineList(100, "")
	suite.NoError(err)
	suite.NotNil(tasks)
	for _, task := range tasks.Tasks {
		err = suite.client.OfflineRetry(task.ID)
		suite.NoError(err)
	}
}

func (suite *TestPikpakSuite) TestTaskRemove() {
	tasks, err := suite.client.OfflineList(100, "")
	suite.NoError(err)
	suite.NotNil(tasks)
	for _, task := range tasks.Tasks {
		if task.Phase == pikpakgo.PhaseTypeComplete {
			err = suite.client.OfflineRemove([]string{task.ID}, true)
			suite.NoError(err)
			break
		}
	}
}

func (suite *TestPikpakSuite) TestBatchTrashFiles() {
	err := suite.client.BatchTrashFiles([]string{
		"VNV9ua9L2OQzryfULN72j50to1",
		"VNVDm8wqQjBlpj7t6p3E9wsMo1",
	})
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestBatchDeleteFiles() {
	err := suite.client.BatchDeleteFiles([]string{
		"VNV9ua9L2OQzryfULN72j50to1",
	})
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestBatchMoveFiles() {
	f1, err := suite.client.CreateFolder("f1", "")
	suite.NoError(err)
	suite.NotNil(f1)
	defer func() {
		err = suite.client.BatchDeleteFiles([]string{f1.ID})
		suite.NoError(err)
	}()
	f2, err := suite.client.CreateFolder("f2", "")
	suite.NoError(err)
	suite.NotNil(f2)
	defer func() {
		err = suite.client.BatchDeleteFiles([]string{f2.ID})
		suite.NoError(err)
	}()

	err = suite.client.BatchMoveFiles([]string{f2.ID}, f1.ID)
	suite.NoError(err)
	f2, err = suite.client.GetFile(f2.ID)
	suite.NoError(err)
	suite.NotNil(f2)
	suite.Equal(f1.ID, f2.ParentID)
}

func (suite *TestPikpakSuite) TestGetAbout() {
	info, err := suite.client.About()
	suite.NoError(err)
	suite.NotNil(info)
}

func (suite *TestPikpakSuite) TestGetMe() {
	info, err := suite.client.Me()
	suite.NoError(err)
	suite.NotNil(info)
}

func (suite *TestPikpakSuite) TestGetInviteInfo() {
	info, err := suite.client.InviteInfo()
	suite.NoError(err)
	suite.NotNil(info)
}

func (suite *TestPikpakSuite) TestCreateFolder() {
	f, err := suite.client.CreateFolder("test5", "")
	suite.NoError(err)
	suite.NotNil(f)
	defer func() {
		err = suite.client.BatchDeleteFiles([]string{f.ID})
		suite.NoError(err)
	}()
	newFile, err := suite.client.RenameFile(f.ID, "renamed")
	suite.NoError(err)
	suite.NotNil(newFile)
	suite.Equal(newFile.ID, f.ID)
	suite.Equal(newFile.Name, "renamed")
}

func (suite *TestPikpakSuite) TestPathToID() {
	testPath := "/home/test"
	id, err := suite.client.FolderPathToID(testPath, true)
	suite.NoError(err)
	suite.NotEmpty(id)
	exists, err := suite.client.FileExists("/home/test")
	suite.NoError(err)
	suite.True(exists)
	err = suite.client.RemoveFolder("/home")
	suite.NoError(err)
	exists, err = suite.client.FileExists("/home")
	suite.NoError(err)
	suite.False(exists)
}

func (suite *TestPikpakSuite) TestOfflineRemoveAll() {
	err := suite.client.OfflineRemoveAll([]string{pikpakgo.PhaseTypeError, pikpakgo.PhaseTypePending, pikpakgo.PhaseTypeRunning}, true)
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestDeleteAllFiles() {
	err := suite.client.OfflineRemoveAll([]string{pikpakgo.PhaseTypeError, pikpakgo.PhaseTypePending, pikpakgo.PhaseTypeRunning}, true)
	suite.NoError(err)
	files, err := suite.client.FileListAll("")
	suite.NoError(err)
	var ids []string
	for _, fi := range files {
		if fi.FolderType == pikpakgo.FolderTypeDownload {
			continue
		}
		ids = append(ids, fi.ID)
	}
	err = suite.client.BatchDeleteFiles(ids)
	suite.NoError(err)
	err = suite.client.EmptyTrash()
	suite.NoError(err)
}

func (suite *TestPikpakSuite) TestGetAllFiles() {
	if getAllFileFieldId == "" {
		return
	}
	files, err := suite.client.FileListAll(getAllFileFieldId)
	suite.NoError(err)
	suite.NotEmpty(files)
	// Make sure your field contains 100+ files
	suite.Greater(len(files), 100)
}
