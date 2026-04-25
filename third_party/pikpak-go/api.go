package pikpakgo

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	ClientId        = "YNxT9w7GMdWvEOKa"
	ClientSecret    = "dbw2OtmVEeuUvIptb1Coyg"
	GrantType       = "password"
	PikpakUserHost  = "https://user.mypikpak.com"
	PikpakDriveHost = "https://api-drive.mypikpak.com"
	PackageName     = `com.pikcloud.pikpak`
	ClientVersion   = `1.47.1`

	// AlgoObjectsString this is information from pikpak website js file. searching keyword: calculateCaptchaSign
	// https://mypikpak.com/drive/main.e3f02a36.js
	AlgoObjectsString = `
	[{
		"alg": "md5",
		"salt": "Gez0T9ijiI9WCeTsKSg3SMlx"
	}, {
		"alg": "md5",
		"salt": "zQdbalsolyb1R/"
	}, {
		"alg": "md5",
		"salt": "ftOjr52zt51JD68C3s"
	}, {
		"alg": "md5",
		"salt": "yeOBMH0JkbQdEFNNwQ0RI9T3wU/v"
	}, {
		"alg": "md5",
		"salt": "BRJrQZiTQ65WtMvwO"
	}, {
		"alg": "md5",
		"salt": "je8fqxKPdQVJiy1DM6Bc9Nb1"
	}, {
		"alg": "md5",
		"salt": "niV"
	}, {
		"alg": "md5",
		"salt": "9hFCW2R1"
	}, {
		"alg": "md5",
		"salt": "sHKHpe2i96"
	}, {
		"alg": "md5",
		"salt": "p7c5E6AcXQ/IJUuAEC9W6"
	}, {
		"alg": "md5",
		"salt": ""
	}, {
		"alg": "md5",
		"salt": "aRv9hjc9P+Pbn+u3krN6"
	}, {
		"alg": "md5",
		"salt": "BzStcgE8qVdqjEH16l4"
	}, {
		"alg": "md5",
		"salt": "SqgeZvL5j9zoHP95xWHt"
	}, {
		"alg": "md5",
		"salt": "zVof5yaJkPe3VFpadPof"
	}]
`
)

var (
	AlgoObjects []AlgoObject
)

type AlgoObject struct {
	Alg  string `json:"alg"`
	Salt string `json:"salt"`
}

func init() {
	err := json.Unmarshal([]byte(AlgoObjectsString), &AlgoObjects)
	if err != nil {
		panic(err)
	}
}

type PikPakClient struct {
	username     string
	password     string
	accessToken  string
	refreshToken string
	sub          string
	tokenAuth    bool
	captchaToken string
	deviceId     string
	client       *resty.Client
}

func NewPikPakClient(username, password string) (*PikPakClient, error) {
	client := resty.New()
	client.EnableTrace()
	client.SetRetryCount(5)
	client.SetRetryWaitTime(5 * time.Second)
	client.SetRetryMaxWaitTime(60 * time.Second)
	client.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36")

	deviceId := md5.Sum([]byte(username))
	pikpak := PikPakClient{
		username: username,
		password: password,
		client:   client,
		deviceId: hex.EncodeToString(deviceId[:]),
	}

	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if r != nil && strings.Contains(string(r.Body()), "unauthenticated") {
			var authErr error
			if pikpak.refreshToken != "" {
				authErr = pikpak.RefreshAccessToken()
				if authErr != nil {
					authErr = pikpak.loginWithPassword()
				}
			} else {
				authErr = pikpak.loginWithPassword()
			}
			if authErr != nil {
				return false
			}
			r.Request.SetAuthToken(pikpak.accessToken)
			return true
		}
		if err == nil {
			return false
		}
		if err != nil {
			return true
		}
		return false
	})

	return &pikpak, nil
}

func (c *PikPakClient) SetTokens(accessToken, refreshToken string) {
	c.accessToken = strings.TrimSpace(accessToken)
	c.refreshToken = strings.TrimSpace(refreshToken)
	c.tokenAuth = c.accessToken != "" || c.refreshToken != ""
}

func (c *PikPakClient) Tokens() (accessToken, refreshToken string) {
	return c.accessToken, c.refreshToken
}

func (c *PikPakClient) Login() error {
	if c.tokenAuth {
		if c.accessToken != "" {
			return nil
		}
		if c.refreshToken != "" {
			if err := c.RefreshAccessToken(); err == nil {
				return nil
			}
		}
	}
	return c.loginWithPassword()
}

func (c *PikPakClient) loginWithPassword() error {
	loginURL := fmt.Sprintf("%s/v1/auth/signin", PikpakUserHost)
	captchaMeta := CaptchaMeta{}
	if strings.Contains(c.username, "@") {
		captchaMeta.Email = c.username
	} else {
		captchaMeta.Username = c.username
	}
	if err := c.CaptchaTokenWithMeta("POST:"+loginURL, captchaMeta); err != nil {
		return err
	}
	req := RequestLogin{
		ClientId:     ClientId,
		ClientSecret: ClientSecret,
		Username:     c.username,
		Password:     c.password,
		CaptchaToken: c.captchaToken,
	}
	resp := ResponseLogin{}
	originResp, err := c.client.R().
		SetBody(&req).
		SetResult(&resp).
		SetHeader("Content-Type", "application/json; charset=utf-8").
		Post(loginURL)
	if err != nil {
		return err
	}
	if resp.AccessToken == "" {
		return errRespToError(originResp.Body())
	}
	c.accessToken = resp.AccessToken
	c.refreshToken = resp.RefreshToken
	c.sub = resp.Sub
	c.tokenAuth = c.accessToken != "" || c.refreshToken != ""
	return nil
}

func (c *PikPakClient) RefreshAccessToken() error {
	if c.refreshToken == "" {
		return errors.New("pikpak refresh token is empty")
	}
	refreshURL := fmt.Sprintf("%s/v1/auth/token", PikpakUserHost)
	req := RequestRefreshToken{
		ClientId:     ClientId,
		RefreshToken: c.refreshToken,
		GrantType:    "refresh_token",
	}
	resp := ResponseLogin{}
	originResp, err := c.client.R().
		SetBody(&req).
		SetResult(&resp).
		SetHeader("Content-Type", "application/json; charset=utf-8").
		Post(refreshURL)
	if err != nil {
		return err
	}
	if resp.AccessToken == "" {
		return errRespToError(originResp.Body())
	}
	c.accessToken = resp.AccessToken
	c.refreshToken = resp.RefreshToken
	c.sub = resp.Sub
	c.tokenAuth = true
	return nil
}

func (c *PikPakClient) Logout() error {
	req := RequestLogout{
		Token: c.accessToken,
	}
	_, err := c.client.R().
		SetBody(&req).
		Post(fmt.Sprintf("%s/v1/auth/revoke", PikpakUserHost))
	return err
}

func (c *PikPakClient) FileList(limit int, parentId string, nextPageToken string) (*FileList, error) {
	err := c.CaptchaToken("GET:/drive/v1/files/")
	if err != nil {
		return nil, err
	}
	filters := Filters{
		Phase: map[string]string{
			"eq": PhaseTypeComplete,
		},
		Trashed: map[string]bool{
			"eq": false,
		},
	}
	filtersBz, err := json.Marshal(&filters)
	if err != nil {
		return nil, err
	}
	req := RequestFileList{
		ParentId:      parentId,
		ThumbnailSize: ThumbnailSizeM,
		Limit:         strconv.Itoa(limit),
		WithAudit:     strconv.FormatBool(true),
		PageToken:     nextPageToken,
		Filters:       string(filtersBz),
	}
	bz, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	reqParams := make(map[string]string)
	err = json.Unmarshal(bz, &reqParams)
	if err != nil {
		return nil, err
	}

	result := FileList{}
	_, err = c.client.R().
		SetQueryParams(reqParams).
		SetResult(&result).
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Get(fmt.Sprintf("%s/drive/v1/files", PikpakDriveHost))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *PikPakClient) FileListAll(fileId string) ([]*File, error) {
	pageSize := 100
	nextPageToken := ""
	var files []*File
	for {
		ls, err := c.FileList(pageSize, fileId, nextPageToken)
		if err != nil {
			return nil, err
		}
		files = append(files, ls.Files...)
		if len(ls.Files) < pageSize || ls.NextPageToken == "" {
			break
		}
		nextPageToken = ls.NextPageToken
	}
	return files, nil
}

func (c *PikPakClient) GetFile(id string) (*File, error) {
	err := c.CaptchaToken("GET:/drive/v1/files/")
	if err != nil {
		return nil, err
	}
	file := File{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetResult(&file).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Get(fmt.Sprintf("%s/drive/v1/files/%s?usage=FETCH", PikpakDriveHost, id))
	if err != nil {
		return nil, err
	}
	return &file, errRespToError(resp.Body())
}

func (c *PikPakClient) CreateFolder(name string, parentId string) (*File, error) {
	params := map[string]string{
		"kind":      KindOfFolder,
		"name":      name,
		"parent_id": parentId,
	}
	err := c.CaptchaToken("POST:/drive/v1/files")
	if err != nil {
		return nil, err
	}
	var result NewFile
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&params).
		SetResult(&result).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Post(fmt.Sprintf("%s/drive/v1/files", PikpakDriveHost))
	if err != nil {
		return nil, err
	}
	return result.File, errRespToError(resp.Body())
}

func (c *PikPakClient) GetDownloadUrl(id string) (string, error) {
	err := c.CaptchaToken("GET:/drive/v1/files/")
	if err != nil {
		return "", err
	}
	file := File{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetResult(&file).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Get(fmt.Sprintf("%s/drive/v1/files/%s?usage=FETCH", PikpakDriveHost, id))
	if err != nil {
		return "", err
	}
	return file.WebContentLink, errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineDownload(name string, fileUrl string, parentId string) (*NewTask, error) {
	folderType := FolderTypeDownload
	if parentId != "" {
		folderType = ""
	}
	req := RequestNewTask{
		Kind:       KindOfFile,
		Name:       name,
		ParentID:   parentId,
		UploadType: UploadTypeURL,
		URL: &URL{
			URL: fileUrl,
		},
		FolderType: folderType,
	}
	err := c.CaptchaToken("POST:/drive/v1/files")
	if err != nil {
		return nil, err
	}
	task := NewTask{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		SetResult(&task).
		SetBody(&req).
		Post(fmt.Sprintf("%s/drive/v1/files", PikpakDriveHost))
	if err != nil {
		return nil, err
	}
	return &task, errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineList(limit int, nextPageToken string) (*TaskList, error) {
	filters := Filters{
		Phase: map[string]string{
			"in": strings.Join([]string{PhaseTypeRunning, PhaseTypeComplete, PhaseTypeError}, ","),
		},
	}
	filtersBz, err := json.Marshal(&filters)
	if err != nil {
		return nil, err
	}
	req := RequestTaskList{
		ThumbnailSize: ThumbnailSizeS,
		Limit:         strconv.Itoa(limit),
		NextPageToken: nextPageToken,
		Filters:       string(filtersBz),
		FileType:      FileTypeOffline,
	}
	bz, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	reqParams := make(map[string]string)
	err = json.Unmarshal(bz, &reqParams)
	if err != nil {
		return nil, err
	}

	err = c.CaptchaToken("GET:/drive/v1/tasks")
	if err != nil {
		return nil, err
	}

	result := TaskList{}
	resp, err := c.client.R().
		SetQueryParams(reqParams).
		SetResult(&result).
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Get(fmt.Sprintf("%s/drive/v1/tasks", PikpakDriveHost))
	if err != nil {
		return nil, err
	}
	return &result, errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineRetry(taskId string) error {
	err := c.CaptchaToken("GET:/drive/v1/task")
	if err != nil {
		return err
	}
	req := RequestTaskRetry{
		Id:         taskId,
		Type:       FileTypeOffline,
		CreateType: CreateTypeRetry,
	}
	bz, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	reqParams := make(map[string]string)
	err = json.Unmarshal(bz, &reqParams)
	if err != nil {
		return err
	}
	resp, err := c.client.R().
		SetQueryParams(reqParams).
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Get(fmt.Sprintf("%s/drive/v1/task", PikpakDriveHost))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineRemove(taskId []string, deleteFiles bool) error {
	err := c.CaptchaToken("DELETE:/drive/v1/tasks")
	if err != nil {
		return err
	}
	params := map[string]string{
		"task_ids":     strings.Join(taskId, ","),
		"delete_files": strconv.FormatBool(deleteFiles),
	}
	resp, err := c.client.R().
		SetQueryParams(params).
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Delete(fmt.Sprintf("%s/drive/v1/tasks", PikpakDriveHost))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) OfflineListIterator(callback func(task *Task) bool) error {
	nextPageToken := ""
	pageSize := 10000
Exit:
	for {
		taskList, err := c.OfflineList(pageSize, nextPageToken)
		if err != nil {
			return err
		}
		for _, task := range taskList.Tasks {
			if callback(task) {
				break Exit
			}
		}
		if len(taskList.Tasks) < pageSize || taskList.NextPageToken == "" {
			break Exit
		}
		nextPageToken = taskList.NextPageToken
	}
	return nil
}

// OfflineRemoveAll remove offline tasks
//   - phase
//     PhaseTypeError...
func (c *PikPakClient) OfflineRemoveAll(phases []string, deleteFiles bool) error {
	var taskIds []string
	err := c.OfflineListIterator(func(task *Task) bool {
		found := false
		for _, phase := range phases {
			if task.Phase == phase {
				found = true
				break
			}
		}
		if len(phases) == 0 || found {
			taskIds = append(taskIds, task.ID)
		}
		return false
	})
	if err != nil {
		return err
	}
	if len(taskIds) > 0 {
		return c.OfflineRemove(taskIds, deleteFiles)
	}
	return nil
}

func (c *PikPakClient) WaitForOfflineDownloadComplete(taskId string, timeout time.Duration, progressFn func(*Task)) (*Task, error) {
	finished := false
	var finishedTask *Task
	var lastErr error
	endTime := time.Now().Add(timeout)
	for {
		if finished {
			return finishedTask, nil
		}
		if time.Now().After(endTime) {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, ErrWaitForOfflineDownloadTimeout
		}
		lastErr = c.OfflineListIterator(func(task *Task) bool {
			if task.ID == taskId {
				if progressFn != nil {
					progressFn(task)
				}
				if (task.Phase == PhaseTypeComplete && task.Progress == 100) || task.Phase == PhaseTypeError {
					finished = true
					finishedTask = task
					return true
				}
			}
			return false
		})
		time.Sleep(5 * time.Second)
	}
}

func (c *PikPakClient) BatchTrashFiles(ids []string) error {
	err := c.CaptchaToken("POST:/drive/v1/files:batchTrash")
	if err != nil {
		return err
	}
	req := RequestBatch{
		Ids: ids,
	}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&req).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Post(fmt.Sprintf("%s/drive/v1/files:batchTrash", PikpakDriveHost))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) BatchDeleteFiles(ids []string) error {
	err := c.CaptchaToken("POST:/drive/v1/files:batchDelete")
	if err != nil {
		return err
	}
	req := RequestBatch{
		Ids: ids,
	}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&req).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Post(fmt.Sprintf("%s/drive/v1/files:batchDelete", PikpakDriveHost))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) BatchMoveFiles(ids []string, folderId string) error {
	err := c.CaptchaToken("POST:/drive/v1/files:batchMove")
	if err != nil {
		return err
	}
	req := RequestBatch{
		Ids: ids,
		To: map[string]string{
			"parent_id": folderId,
		},
	}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&req).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Post(fmt.Sprintf("%s/drive/v1/files:batchMove", PikpakDriveHost))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) EmptyTrash() error {
	err := c.CaptchaToken("PATCH:/drive/v1/files/trash:empty")
	if err != nil {
		return err
	}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Patch(fmt.Sprintf("%s/drive/v1/files/trash:empty", PikpakDriveHost))
	if err != nil {
		return err
	}
	return errRespToError(resp.Body())
}

func (c *PikPakClient) RenameFile(id string, name string) (*File, error) {
	err := c.CaptchaToken("PATCH:/drive/v1/files/")
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"name": name,
	}
	file := File{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&params).
		SetResult(&file).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Patch(fmt.Sprintf("%s/drive/v1/files/%s", PikpakDriveHost, id))
	if err != nil {
		return nil, err
	}
	return &file, errRespToError(resp.Body())
}

func (c *PikPakClient) About() (*About, error) {
	err := c.CaptchaToken("GET:/drive/v1/about")
	if err != nil {
		return nil, err
	}
	info := About{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetResult(&info).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		Get(fmt.Sprintf("%s/drive/v1/about", PikpakDriveHost))
	if err != nil {
		return &info, err
	}
	return &info, errRespToError(resp.Body())
}

func (c *PikPakClient) CaptchaToken(action string) error {
	return c.CaptchaTokenWithMeta(action, CaptchaMeta{})
}

func (c *PikPakClient) CaptchaTokenWithMeta(action string, meta CaptchaMeta) error {
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())
	sign := ClientId + ClientVersion + PackageName + c.deviceId + ts
	for _, algo := range AlgoObjects {
		sign = fmt.Sprintf("%x", md5.Sum([]byte(sign+algo.Salt)))
	}
	sign = "1." + sign
	meta.CaptchaSign = sign
	meta.ClientVersion = ClientVersion
	meta.PackageName = PackageName
	meta.Timestamp = ts
	if meta.UserID == "" {
		meta.UserID = c.sub
	}
	req := RequestGetCaptcha{
		Action:   action,
		ClientID: ClientId,
		DeviceID: c.deviceId,
		Meta:     meta,
	}
	resp := ResponseGetCaptcha{}
	originResp, err := c.client.R().
		SetQueryParam("client_id", ClientId).
		SetResult(&resp).
		SetBody(&req).
		Post(fmt.Sprintf("%s/v1/shield/captcha/init", PikpakUserHost))
	if err != nil {
		return err
	}
	c.captchaToken = resp.CaptchaToken
	return errRespToError(originResp.Body())
}

func (c *PikPakClient) Me() (*MeInfo, error) {
	err := c.CaptchaToken("GET:/v1/user/me")
	if err != nil {
		return nil, err
	}
	result := MeInfo{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		SetResult(&result).
		Get(fmt.Sprintf("%s/v1/user/me", PikpakUserHost))
	if err != nil {
		return nil, err
	}
	return &result, errRespToError(resp.Body())
}

func (c *PikPakClient) InviteInfo() (*InviteInfo, error) {
	err := c.CaptchaToken("POST:/vip/v1/activity/invite")
	if err != nil {
		return nil, err
	}
	req := map[string]string{
		"from": "web",
	}
	result := InviteInfo{}
	resp, err := c.client.R().
		SetAuthToken(c.accessToken).
		SetBody(&req).
		SetHeader("x-captcha-token", c.captchaToken).
		SetHeader("x-device-id", c.deviceId).
		SetResult(&result).
		Post(fmt.Sprintf("%s/vip/v1/activity/invite", PikpakDriveHost))
	if err != nil {
		return nil, err
	}
	return &result, errRespToError(resp.Body())
}

func (c *PikPakClient) WalkDir(fileId string, fn func(file *File) bool) error {
	rootFile, err := c.GetFile(fileId)
	if err != nil {
		return err
	}
	stack := []*FileTreeNode{
		{
			Paths: []string{},
			File:  rootFile,
		},
	}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if fn(node.File) {
			return nil
		}
		if node.File.Kind == KindOfFolder {
			fileList, err := c.FileListAll(node.File.ID)
			if err != nil {
				return err
			}
			// using for callback
			for _, f := range fileList {
				stack = append(stack, &FileTreeNode{
					Paths: append(node.Paths, node.File.Name),
					File:  f,
				})
			}
		}
	}
	return nil
}

func (c *PikPakClient) FileExists(absPath string) (bool, error) {
	if !filepath.IsAbs(absPath) {
		return false, errors.New("path is not absolute")
	}
	if absPath == "/" {
		return true, nil
	}

	dirNames := strings.Split(absPath, "/")
	dirId := ""

	for _, dirName := range dirNames {
		if dirName == "" {
			continue
		}
		found := false
		files, err := c.FileListAll(dirId)
		if err != nil {
			return false, err
		}
		for _, fi := range files {
			if fi.Name == dirName {
				found = true
				dirId = fi.ID
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}

// FolderPathToID the path format: /home/test
func (c *PikPakClient) FolderPathToID(absPath string, autoCreate bool) (string, error) {
	if !filepath.IsAbs(absPath) {
		return "", errors.New("path is not absolute")
	}
	if absPath == "/" {
		return "", nil
	}

	dirNames := strings.Split(absPath, "/")
	dirId := ""

	for _, dirName := range dirNames {
		if dirName == "" {
			continue
		}
		dirFound := false
		files, err := c.FileListAll(dirId)
		if err != nil {
			return "", err
		}
		for _, fi := range files {
			if fi.Name == dirName {
				if fi.Kind == KindOfFolder {
					dirFound = true
					dirId = fi.ID
					break
				} else {
					return "", fmt.Errorf("the dir %s is file, not a dir", dirName)
				}
			}
		}
		if !dirFound && autoCreate {
			newDir, err := c.CreateFolder(dirName, dirId)
			if err != nil {
				return "", err
			}
			dirId = newDir.ID
		}
		if !dirFound && !autoCreate {
			return "", ErrFolderNotFound
		}
	}
	return dirId, nil
}

func (c *PikPakClient) RemoveFolder(absPath string) error {
	id, err := c.FolderPathToID(absPath, false)
	if errors.Is(err, ErrFolderNotFound) {
		return nil
	}
	return c.BatchDeleteFiles([]string{id})
}

func errRespToError(body []byte) error {
	pikpakErr := Error{}
	err := json.Unmarshal(body, &pikpakErr)
	if err != nil {
		return nil
	} else if pikpakErr.Code != 0 && pikpakErr.Reason != "" {
		switch pikpakErr.Reason {
		case "file_space_not_enough":
			return ErrSpaceNotEnough
		case "task_daily_create_limit":
			return ErrDailyCreateLimit
		case "file_duplicated_name":
			return ErrFileDuplicateName
		}
		return errors.New(pikpakErr.Error())
	}
	return nil
}
