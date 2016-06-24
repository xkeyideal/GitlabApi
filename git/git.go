package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"util"
	"v1/leonid/config"
)

/*------------------------------------------------------------*/
/*-------------------Git API----------------------------------*/
/*------------------------------------------------------------*/

func GitCloneToDir(remoteUrl, projectName string) (msg string, err error) {
	if err = util.PingRemote(strings.Split(remoteUrl, ":")[0]); err != nil {
		err = util.NewError(fmt.Sprintf("Cannot SSH To Remote. Check ssh-key : %s", err.Error()))
		return
	}

	deployDir := config.GitDeployDir + projectName

	if ok, _ := util.IsDir(deployDir); ok {
		err = util.NewError("Project[%s] Directory Existed", projectName)
		return
	}

	err = os.MkdirAll(path.Dir(deployDir), 0700)
	if err != nil {
		err = util.NewError("Project[%s] Directory Mkdir Failed: %s", projectName, err.Error())
		return
	}

	so, se, err := util.RunCmd(config.GIT, "clone", remoteUrl, deployDir)

	if len(se) != 0 {
		err = util.NewError("Command Git Clone exec stderr: %s", se)
		return
	}

	msg = so

	return
}

/*
git-pull command 不支持 --work-tree参数，因此必须使用
git fetch origin
git merge origin/master
代替，本身merge的过程中也可能出现merge出错

git remote set-url origin remoteUrl
git checkout branchname
git fetch origin branchname
git merge origin/branchname
*/
func GitPullToDir(remoteUrl, projectName, branchName string) (msg string, err error) {
	if err = util.PingRemote(strings.Split(remoteUrl, ":")[0]); err != nil {
		err = util.NewError(fmt.Sprintf("Cannot SSH To Remote. Check ssh-key : %s", err.Error()))
		return
	}

	deployDir := config.GitDeployDir + projectName

	if ok, _ := util.IsDir(deployDir); !ok {
		err = util.NewError("Project[%s] Git Repository Not Existed", projectName)
		return
	}

	gitDir := path.Join(deployDir, ".git")

	args := [][]string{
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s remote set-url origin %s", gitDir, deployDir, remoteUrl), " "),
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s checkout %s", gitDir, deployDir, branchName), " "),
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s fetch origin %s", gitDir, deployDir, branchName), " "),
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s merge origin/%s", gitDir, deployDir, branchName), " "),
	}

	for _, arg := range args {
		so, se, e := util.RunCmd(config.GIT, arg...)
		msg += (so + se)
		if e != nil {
			err = e
			return
		}

		//		if len(se) != 0 {
		//			err = util.NewError("Command Git Pull exec stderr: %s", se)
		//			return
		//		}
	}
	return
}

//通过读取文件获取项目最新的commitID
func GetGitHeadCommitid(projectName string) (commitId string, err error) {
	deployDir := config.GitDeployDir + projectName

	if ok, _ := util.IsDir(deployDir); !ok {
		err = util.NewError("Project[%s] Directory Not Existed", projectName)
		return
	}

	headPath := path.Join(deployDir, ".git/HEAD")
	content, err := ioutil.ReadFile(headPath)

	if err != nil {
		return
	}

	re, err := regexp.Compile(`ref:\s([\S]+)`)
	if err != nil {
		return
	}

	if len(re.FindStringSubmatch(string(content))) == 2 {
		headPath = path.Join(deployDir, ".git/", re.FindStringSubmatch(string(content))[1])
		var e error
		content, e = ioutil.ReadFile(headPath)
		if e != nil {
			err = e
			return
		}
	} else {
		err = util.NewError("Get Git Head CommitId Failed, regex match error")
		return
	}
	commitId = strings.TrimSpace(string(content))
	return
}

//通过读取文件获取某个branch最新的commitID
func GetGitBranchHeadCommitid(projectName, branchName string) (commitId string, err error) {
	deployDir := config.GitDeployDir + projectName

	if ok, _ := util.IsDir(deployDir); !ok {
		err = util.NewError("Project[%s] Directory Not Existed", projectName)
		return
	}

	headPath := path.Join(deployDir, fmt.Sprintf(".git/refs/heads/%s", branchName))

	if ok := util.IsExist(headPath); !ok {
		err = util.NewError("[%s] Branch Head File Not Exist, Please Make Sure The Branch Existed", branchName)
		return
	}

	content, err := ioutil.ReadFile(headPath)

	if err != nil {
		return
	}
	commitId = strings.TrimSpace(string(content))
	return
}

/*
	git remote set-url origin remoteUrl
	git checkout branchname
	git add -A
	git commit -am commitmessage
	git push origin branchname
*/
func GitPushToRemote(remoteUrl, projectName, branchName, commitMsg string) (msg string, err error) {

	if ok := strings.Contains(commitMsg, "|||"); ok {
		err = util.NewError("Commit Message Can't Contains [|||] String")
		return
	}

	if err = util.PingRemote(strings.Split(remoteUrl, ":")[0]); err != nil {
		err = util.NewError(fmt.Sprintf("Cannot SSH To Remote. Check ssh-key : %s", err.Error()))
		return
	}

	deployDir := config.GitDeployDir + projectName

	if ok, _ := util.IsDir(deployDir); !ok {
		err = util.NewError("Project[%s] Directory Not Existed", projectName)
		return
	}

	gitDir := path.Join(deployDir, ".git")

	args := [][]string{
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s remote set-url origin %s", gitDir, deployDir, remoteUrl), " "),
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s checkout %s", gitDir, deployDir, branchName), " "),
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s add -A", gitDir, deployDir), " "),
		strings.Split(fmt.Sprintf("--git-dir=%s|||--work-tree=%s|||commit|||-am|||%s", gitDir, deployDir, commitMsg), "|||"),
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s push origin %s", gitDir, deployDir, branchName), " "),
	}

	for _, arg := range args {
		so, se, e := util.RunCmd(config.GIT, arg...)
		msg += (so + se)
		if e != nil {
			err = e
			return
		}
	}
	return
}

//获取项目指定branch的commit信息，包括id和message
func GitFileCommitids(projectName, branchName, filepath string) (commitInfos []CommitInfo, err error) {
	deployDir := config.GitDeployDir + projectName

	gitDir := path.Join(deployDir, ".git")

	if ok, _ := util.IsDir(deployDir); !ok {
		err = util.NewError("Project[%s] Directory Not Existed", projectName)
		return
	}

	realFilePath := path.Join(deployDir, filepath)

	if ok := util.IsExist(realFilePath); !ok {
		err = util.NewError("File[%s] Not Exist, Please Check", filepath)
		return
	}

	args := [][]string{
		strings.Split(fmt.Sprintf("--git-dir=%s --work-tree=%s log --pretty=format:\"%%H:%%s\" %s %s", gitDir, deployDir, branchName, realFilePath), " "),
	}

	msg := ""
	for _, arg := range args {
		so, se, e := util.RunCmd(config.GIT, arg...)
		msg += so
		if e != nil {
			err = e
			return
		}

		if len(se) != 0 {
			err = util.NewError("Command Git log exec stderr: %s", se)
			return
		}
	}

	infos := strings.Split(msg, "\n")

	for _, info := range infos {
		ss := strings.Split(strings.Trim(info, "\""), ":")
		if len(ss) != 2 {
			err = util.NewError("Commit Info [%s] Split Error", info)
			return
		}
		commitInfos = append(commitInfos, CommitInfo{Id: ss[0], Msg: ss[1]})
	}

	return
}
