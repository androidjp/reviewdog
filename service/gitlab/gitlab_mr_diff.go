package gitlab

import (
	"context"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"os/exec"

	"github.com/androidjp/reviewdog"
	"github.com/androidjp/reviewdog/service/serviceutil"
)

var _ reviewdog.DiffService = &MergeRequestDiff{}

// MergeRequestDiff is a diff service for GitLab MergeRequest.
type MergeRequestDiff struct {
	cli      *gitlab.Client
	pr       int
	sha      string
	projects string

	// wd is working directory relative to root of repository.
	wd string
}

// NewGitLabMergeRequestDiff returns a new MergeRequestDiff service.
// itLabMergeRequestDiff service needs git command in $PATH.
func NewGitLabMergeRequestDiff(cli *gitlab.Client, owner, repo string, pr int, sha string) (*MergeRequestDiff, error) {
	workDir, err := serviceutil.GitRelWorkdir()
	if err != nil {
		return nil, fmt.Errorf("MergeRequestCommitCommenter needs 'git' command: %w", err)
	}
	return &MergeRequestDiff{
		cli:      cli,
		pr:       pr,
		sha:      sha,
		projects: owner + "/" + repo,
		wd:       workDir,
	}, nil
}

// Diff returns a diff of MergeRequest. It runs `git diff` locally instead of
// diff_url of GitLab Merge Request because diff of diff_url is not suited for
// comment API in a sense that diff of diff_url is equivalent to
// `git diff --no-renames`, we want diff which is equivalent to
// `git diff --find-renames`.
func (g *MergeRequestDiff) Diff(ctx context.Context) ([]byte, error) {
	mr, _, err := g.cli.MergeRequests.GetMergeRequest(g.projects, g.pr, nil)
	if err != nil {
		return nil, err
	}
	targetBranch, _, err := g.cli.Branches.GetBranch(mr.TargetProjectID, mr.TargetBranch, nil)
	if err != nil {
		return nil, err
	}
	return g.gitDiff(ctx, g.sha, targetBranch.Commit.ID)
}

func (g *MergeRequestDiff) gitDiff(_ context.Context, baseSha, targetSha string) ([]byte, error) {
	//var out bytes.Buffer
	//var stderr bytes.Buffer
	//cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("git merge-base %s %s", targetSha, baseSha))
	//cmd.Stdout = &out
	//cmd.Stderr = &stderr
	//err := cmd.Run()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get merge-base commit with detail2: %s, baseSHA:%s, targetSHA: %s, stdout: %s", fmt.Sprint(err)+": "+stderr.String(), baseSha, targetSha, out.String())
	//}
	//mergeBase := strings.Trim(out.String(), "\n")
	bs, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("git diff --find-renames %s %s", targetSha, baseSha)).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff: %w", err)
	}
	return bs, nil
}

// Strip returns 1 as a strip of git diff.
func (g *MergeRequestDiff) Strip() int {
	return 1
}
