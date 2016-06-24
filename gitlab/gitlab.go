package gitlab

import (
	//"encoding/json"

	"util"
	//"time"

	"config"

	"github.com/astaxie/beego/httplib"
	"github.com/bitly/go-simplejson"
	//"github.com/smallnest/goreq"
)

type ProjectInfo struct {
	ProjectId    int    `json:"id"`
	SshUrlToRepo string `json:"ssh_url_to_repo"`
}

type ProjectBranchInfo struct {
	Name            string
	CommitId        string
	CommitMsg       string
	ParentCommitIds []string
}

type RepoFile struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Size     int    `json:"size"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	Ref      string `json:"ref"`
	CommitId string `json:"commit_id"`
}

type RepoTree struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Mode string `json:"mode"`
	Id   string `json:"id"`
}

type RepoUpdateFile struct {
	FilePath   string `json:"file_path"`
	BranchName string `json:"branch_name"`
}

type CommitInfo struct {
	Id  string
	Msg string
}

type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PrivateToken string `json:"private_token"`
}

/*
默认连接超时和读写超时都使用beego默认值60秒
*/

//通过AdminToken获取当前用户username的信息,暂时没有管理员权限，无法使用
func GitUserAuth(username string) (user User, err error) {
	auth_url := config.GitUrl + config.APIVersion + "/user"

	req := httplib.Get(auth_url)
	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)
	req.Header("SUDO", username)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&user)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//创建一个新的Project，统一创建在[slnanal] namespace下面
func CreateProject(projectName string) (statusCode int, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects"

	req := httplib.Post(project_url)
	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("name", projectName)
	req.Param("namespace_id", config.NamespaceId)
	req.Param("public", "false")

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		statusCode = resp.StatusCode
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//更新项目的名称和Path
func UpdateProject(projectId, newProjectName string) (statusCode int, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId

	req := httplib.Put(project_url)
	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("name", newProjectName)
	req.Param("path", newProjectName)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		statusCode = resp.StatusCode
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//通过项目的namespace和name查询项目信息
func SearchProjectByName(namespace, projectName string) (projectInfo ProjectInfo, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + namespace + "%2F" + projectName

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&projectInfo)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//通过项目ID查询项目信息
func SearchProjectById(projectId string) (projectInfo ProjectInfo, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&projectInfo)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//通过项目名称获取branch的信息
func ListProjectBranchInfoByName(namespace, projectName, branchName string) (projectBranchInfo *ProjectBranchInfo, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + namespace + "%2F" + projectName + "/repository/branches/" + branchName

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	resp, err := req.Response()

	if err != nil {
		return
	}

	projectBranchInfo = &ProjectBranchInfo{}

	if resp.StatusCode == 200 {
		jsonBytes, e := req.Bytes()
		if e != nil {
			err = e
			return
		}

		js, e := simplejson.NewJson(jsonBytes)
		if e != nil {
			err = e
			return
		}

		projectBranchInfo.Name = js.Get("name").MustString()
		projectBranchInfo.CommitId = js.Get("commit").Get("id").MustString()
		projectBranchInfo.CommitMsg = js.Get("commit").Get("message").MustString()
		projectBranchInfo.ParentCommitIds = js.Get("commit").Get("parent_ids").MustStringArray()
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//通过项目ID获取branch的信息
func ListProjectBranchInfoById(projectId, branchName string) (projectBranchInfo *ProjectBranchInfo, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/branches/" + branchName

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	resp, err := req.Response()

	if err != nil {
		return
	}

	projectBranchInfo = &ProjectBranchInfo{}

	if resp.StatusCode == 200 {
		jsonBytes, e := req.Bytes()
		if e != nil {
			err = e
			return
		}

		js, e := simplejson.NewJson(jsonBytes)
		if e != nil {
			err = e
			return
		}

		projectBranchInfo.Name = js.Get("name").MustString()
		projectBranchInfo.CommitId = js.Get("commit").Get("id").MustString()
		projectBranchInfo.CommitMsg = js.Get("commit").Get("message").MustString()
		projectBranchInfo.ParentCommitIds = js.Get("commit").Get("parent_ids").MustStringArray()

	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//获取文件的最新内容,包括content和commit_id等
func GetFileContentRepo(projectId, branchName, filepath string) (repoFile RepoFile, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/files"

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("file_path", filepath)
	req.Param("ref", branchName)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&repoFile)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}
	return
}

//在项目中创建新的文件
func CreateNewFileRepo(projectId, branchName, filepath, content, commitMsg string) (repoUpdateFile RepoUpdateFile, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/files"

	req := httplib.Post(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("file_path", filepath)
	req.Param("branch_name", branchName)
	req.Param("content", content)
	req.Param("encoding", "text")
	req.Param("commit_message", commitMsg)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		err = req.ToJSON(&repoUpdateFile)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}

	return
}

//更新项目中文件的内容
func UpdateExistFileRepo(projectId, branchName, filepath, content, commitMsg string) (repoUpdateFile RepoUpdateFile, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/files"

	req := httplib.Put(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("file_path", filepath)
	req.Param("branch_name", branchName)
	req.Param("content", content)
	req.Param("encoding", "text")
	req.Param("commit_message", commitMsg)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&repoUpdateFile)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}

	return
}

//删除项目中已存在的文件，该功能暂时不能用，会返回400，是权限的问题
func DeleteExistFileRepo(projectId, branchName, filepath, commitMsg string) (repoUpdateFile RepoUpdateFile, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/files"

	req := httplib.Delete(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("file_path", filepath)
	req.Param("branch_name", branchName)
	req.Param("commit_message", commitMsg)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&repoUpdateFile)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}

	return
}

//func DeleteExistFileRepo2(projectId, branchName, filepath, commitMsg string) (repoUpdateFile RepoUpdateFile, err error) {
//	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/files"

//	q := struct {
//		Filepath   string `json:"file_path"`
//		BranchName string `json:"branch_name"`
//		CommitMsg  string `json:"commit_message"`
//	}{
//		Filepath:   filepath,
//		BranchName: branchName,
//		CommitMsg:  commitMsg,
//	}

//	resp, body, errs := goreq.New().Timeout(6000*time.Millisecond).
//		Delete(project_url).
//		ContentType("application/json").
//		SetHeader("PRIVATE-TOKEN", config.AdminToken).SendStruct(q).End()

//	if len(errs) > 0 {
//		err = errs[0]
//		return
//	}

//	if resp.StatusCode == 200 {
//		err = json.Unmarshal(body, &repoUpdateFile)
//		//err = req.ToJSON(&repoUpdateFile)
//	} else {
//		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
//	}

//	return
//}

//根据子目录获取该目录下的文件或子目录信息，不会自动递归子目录查询
func ListRepoTreeByDirectory(projectId, branchName, filepath string) (repoTrees []RepoTree, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/tree"

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("path", filepath)
	req.Param("ref_name", branchName)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&repoTrees)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}

	return
}

//获取项目根目录下的所有子目录和文件信息，不会递归查询
func ListRepoTree(projectId, branchName string) (repoTrees []RepoTree, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/tree"

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("ref_name", branchName)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		err = req.ToJSON(&repoTrees)
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}

	return
}

//根据commitid获取文件的内容
func GetFileContentByCommitid(projectId, sha, filepath string) (content string, err error) {
	project_url := config.GitUrl + config.APIVersion + "/projects/" + projectId + "/repository/blobs/" + sha

	req := httplib.Get(project_url)

	req.Header("Content-Type", "application/json")
	req.Header("PRIVATE-TOKEN", config.AdminToken)

	req.Param("filepath", filepath)

	resp, err := req.Response()

	if err != nil {
		return
	}

	if resp.StatusCode == 200 {
		content, err = req.String()
	} else {
		err = util.NewError("Http Connect Error, Status:%s", resp.Status)
	}

	return
}
