///////////////////////////////////////////////////////////
// git.go
// git api 
// ycyxuehan kun1.huang@outlook.com
//////////////////////////////////////////////////////////


package git

import (
	"time"
	"os"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"strings"
	"fmt"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)
//
const (
	// MixedReset resets the index but not the working tree (i.e., the changed
	// files are preserved but not marked for commit) and reports what has not
	// been updated. This is the default action.
	MixedReset gogit.ResetMode = iota
	// HardReset resets the index and working tree. Any changes to tracked files
	// in the working tree are discarded.
	HardReset
	// MergeReset resets the index and updates the files in the working tree
	// that are different between Commit and HEAD, but keeps those which are
	// different between the index and working tree (i.e. which have changes
	// which have not been added).
	//
	// If a file that is different between Commit and the index has unstaged
	// changes, reset is aborted.
	MergeReset
	// SoftReset does not touch the index file or the working tree at all (but
	// resets the head to <commit>, just like all modes do). This leaves all
	// your changed files "Changes to be committed", as git status would put it.
	SoftReset
)
//Git git
type Git struct {
	Path string
	repo *gogit.Repository
}
type BranchInfo struct{
	Name string `json:"name"`
	Hash string `json:"hash"`
	CommitInfo string `json:"commitinfo"`
	Type string `json:"type"`
	Committer string `json:"committer"`
	CommitTime time.Time `json:"committime"`
}
//New new
func New(path string)(*Git, error){
	if path == "" {
		return nil, fmt.Errorf("git path is empty")
	}
	var g Git
	g.Path = path
	repo, err:= gogit.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	g.repo = repo
	return &g, nil
}

func NewRemote(repo string)(*Git, error){
	if repo == "" {
		return nil, fmt.Errorf("git repo is empty")
	}
	g := Git{
		Path: repo,
	}
	//repo, err := gogit.
	return &g, nil
}

//Reset reset
func (g *Git)Reset(mode gogit.ResetMode)error{
	if g.repo == nil {
		return fmt.Errorf("repo is nil")
	}
	wt, err := g.repo.Worktree()
	if err != nil {
		return err
	}
	opt := gogit.ResetOptions{
		Mode: mode,
	}
	err = wt.Reset(&opt)
	return err
}
//Fetch fetch
func (g *Git)Fetch()error{
	if g.repo == nil {
		return fmt.Errorf("repo is nil")
	}
	opt := gogit.FetchOptions{
		Auth:g.GetSSHAuth(),
		Force: true,
	}
	remote, err := g.GetRemote()
	if err != nil {
		return err
	}
	err = remote.Fetch(&opt)
	return err
}
//Checkout check out a branch or tag
func (g *Git)Checkout(branch string)error{
	if branch == "" {
		return g.Pull()
	}
	if g.repo == nil {
		return fmt.Errorf("repo is nil")
	}
	err := g.Fetch()
	if err != nil {
		if strings.Index(strings.ToLower(err.Error()), "already up-to-date") < 0 {
			return fmt.Errorf("fetch error: %s", err.Error())
		}
	}
	ref, err := g.GetRef(branch)
	if ref == nil {
		return fmt.Errorf("can not get the branch: <%s> in git", branch)
	}
	wt, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("get worktree error: %s", err.Error())
	}
	chkOpt := gogit.CheckoutOptions{
		Branch: ref.Name(),
	}
	err = wt.Checkout(&chkOpt)
	if err != nil {
		return fmt.Errorf("check out branch %s error: %s", branch, err.Error() )
	}
	return nil
}

//GetRef get the branch
func (g *Git)GetRef(branch string)(*plumbing.Reference, error){
	ref, _ := g.Head()
	if branch == ""{
		return ref, nil
	}
	
	branchs, err := g.GetBranchs(true)
	if err != nil {
		return nil, err
	}
	for  _, ref := range branchs{
		if ref.Hash().IsZero(){
			continue
		}
		if ref.Name().Short() == branch {
			// revision := plumbing.Revision(ref.Name().String())
			// commitHash, err := g.repo.ResolveRevision(revision)
			// if err != nil {
			// 	return ref
			// }
			// commit, err := g.repo.CommitObject(*commitHash)
			// if err != nil {
			// 	return ref
			// }

			return ref, nil
		}
		// if ref.Name().IsBranch() && (ref.Name().Short() == fmt.Sprintf("%s/%s", gogit.DefaultRemoteName, branch) || ref.Name().Short() == branch) {
		// 	return ref
		// }
	}
	return nil, fmt.Errorf("not found")
}

//Pull pull the repo
func (g *Git)Pull()error{
	// err := g.Fetch()
	// if err != nil && err.Error() != "already up-to-date"{
	// 	fmt.Println("fetch", err)
	// 	return err
	// }
	head, _ := g.Head()
	opt := gogit.PullOptions{
		Auth: g.GetSSHAuth(),
		RemoteName: gogit.DefaultRemoteName,
		ReferenceName: head.Name(),
	}
	wt, err := g.repo.Worktree()
	if err != nil  {
		return fmt.Errorf("get worktree error: %s", err.Error())
	}
	err = wt.Pull(&opt)
	if err != nil && strings.Index(strings.ToLower(err.Error()), "already up-to-date") < 0 {
		return fmt.Errorf("git pull error: %s", err.Error())
	}
	return nil
}

//Push push the repo
func (g *Git)Push()error{
	if g.repo == nil {
		return fmt.Errorf("repo is nil")
	}
	opt := gogit.PushOptions{}
	return g.repo.Push(&opt)
}

//IsClean work area is clean
func (g *Git)IsClean()(bool, error){
	if g.repo == nil {
		return false, fmt.Errorf("repo is nil")
	}
	wt, err := g.repo.Worktree()
	if err != nil {
		return false, err
	}
	status, err := wt.Status()
	if err != nil {
		return false, err
	}
	return status.IsClean(), nil
}
//IsChanged is the file or path changed
func (g *Git)IsChanged(path string, hash string)(bool, error){
	if g.repo == nil {
		return false, fmt.Errorf("repo is nil")
	}
	commit, err := g.repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return false, err
	}
	files, err := commit.Stats()
	if err != nil {
		return false, err
	}
	for _, file := range files {
		if path == "" {
			return true, nil
		}
		if strings.Index(path, "/") >-1 && strings.Index(file.Name, path) > -1 {
			return true, nil
		} else if file.Name == path{
			return true, nil
		}
	}
	return false, nil
}
//
//GetCurrentRef get current refrence
func (g *Git)GetCurrentRef()(*plumbing.Reference, error){
	if g.repo == nil {
		return nil, fmt.Errorf("repo is nil")
	}
	return g.repo.Head()
}
//Head get current refrence
func (g *Git)Head()(*plumbing.Reference, error){
	if g.repo == nil {
		return nil, fmt.Errorf("repo is nil")
	}
	return g.repo.Head()
}

//GetRemote get remote
func (g *Git)GetRemote()(*gogit.Remote, error){
	remote, err := g.repo.Remote(gogit.DefaultRemoteName)
	return remote, err
}

//GetCommitInfo get git commit info
func (g *Git)GetCommitInfo(path string)string {
	var info = ""
	head, err := g.Head()
	if err != nil {
		return info
	}

	opt := gogit.LogOptions{
		From:head.Hash(),

	}
	log, err := g.repo.Log(&opt)
	if err != nil {
		return info
	}
	c, err := log.Next()
	if err != nil {
		return info
	}
	info = fmt.Sprintf("[%s] %s", c.Committer.Name, strings.Trim(c.Message, "\n"))
	return info
}

//
func (g *Git)GetSSHAuth()transport.AuthMethod{ //*gitssh.PublicKeys{
	s := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	auth, _ := gitssh.NewPublicKeysFromFile("git", s, "")
	return auth
}
//
func (g *Git)GetBranchs(tag bool)([]*plumbing.Reference, error){
	res := []*plumbing.Reference{}
	remotes, err := g.repo.Remotes()
	if err == nil {
		for _, remote := range remotes{
			listOpt := gogit.ListOptions{
				Auth: g.GetSSHAuth(),
			}
			fetchOpt := gogit.FetchOptions{
				Auth: g.GetSSHAuth(),
			}
			remote.Fetch(&fetchOpt)
			list , err :=remote.List(&listOpt)
			if err == nil {
				for _, ref := range list{
					if  ref.Name().IsBranch() || (ref.Name().IsTag() && tag) {
						res = append(res, ref)
					}
				}
			}else{
				return res, err
			}
		}
	}
	return res, err
}
//
func (g *Git)GetBranchsInfo(tag bool)[]BranchInfo{
	res := []BranchInfo{}
	branchs, err := g.GetBranchs(tag)
	if err != nil {
		return res
	}
	for _, ref := range branchs {
		info := BranchInfo{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		}
		info.Type =  "branch"
		opt := gogit.LogOptions{
			From:ref.Hash(),
		}
		if ref.Name().IsTag() {
			revision := plumbing.Revision(ref.Name().String())
			commitHash, err := g.repo.ResolveRevision(revision)
			if err == nil {
				opt.From = *commitHash
			}
			info.Type = "tag"
		}
		log, err := g.repo.Log(&opt)
		if err == nil {
			if c, err := log.Next(); err == nil {
				info.Committer = c.Committer.Name
				info.CommitTime = c.Committer.When
				info.CommitInfo = strings.Trim(c.Message, "\n")
			}
		}
		res = append(res, info)		
	}
	return res
}

//AddTag add a tag
func (g *Git)AddTag(tag string, hash string)error{
	if tag == "" {
		return fmt.Errorf("tag is empty")
	}
	refHash := plumbing.NewHash(hash)
	if hash == "" {
		head, err := g.Head()
		if err != nil {
			return err
		}
		refHash = head.Hash()
	}
	ref := plumbing.NewHashReference(plumbing.ReferenceName(fmt.Sprintf("ref/tags/%s", tag)), refHash)
	return g.repo.Storer.SetReference(ref)
}

func (g *Git)DelTag(tag string)error {
	if tag == "" {
		return fmt.Errorf("tag is empty")
	}
	ref, err := g.GetRef(tag)
	if err != nil {
		return err
	}
	err = g.repo.Storer.RemoveReference(ref.Name())
	return err
}