package libgit

import (
	"bipgit/internal/models"

	git "github.com/libgit2/git2go/v33"
)

func GetLogsOfBranch(repo *git.Repository, branchName string, startCommitID string) ([]*models.Log, string, error) {

	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return nil, "", err
	}

	var startOid *git.Oid
	if startCommitID != "" {
		startOid, _ = git.NewOid(startCommitID)
	} else {
		startOid = branch.Target()
	}

	revWalk, err := repo.Walk()
	revWalk.Sorting(git.SortTime)
	revWalk.Push(startOid)
	if err != nil {
		return nil, "", err
	}

	logs := []*models.Log{}

	oID := startOid
	next := ""
	for {
		err := revWalk.Next(oID)
		if err != nil {
			next = ""
			break
		}
		if len(logs) == 20 {
			next = oID.String()
			break
		}

		commit, _ := repo.LookupCommit(oID)
		message := commit.Message()

		log := models.Log{
			ID:          oID.String(),
			Message:     message,
			AuthorEmail: commit.Author().Email,
			CreatedAt:   commit.Author().When,
		}
		logs = append(logs, &log)
	}

	return logs, next, nil
}
