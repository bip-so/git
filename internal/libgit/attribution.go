package libgit

import (
	"bipgit/internal/models"
	"bipgit/internal/utils"
	"fmt"
	"strings"
	"time"

	git "github.com/libgit2/git2go/v33"
)

func GetBranchAttributions(repo *git.Repository, branchName string) (*[]models.BlockAttribution, error) {

	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return nil, err
	}

	commitTarget, err := repo.LookupCommit(branch.Target())
	if err != nil {
		return nil, err
	}

	commitTree, err := commitTarget.Tree()
	if err != nil {
		return nil, err
	}

	blockIDs := []string{}
	err = commitTree.Walk(func(_ string, treeEntry *git.TreeEntry) error {
		if treeEntry.Name != models.RanksFileName {
			blockIDs = append(blockIDs, strings.ReplaceAll(treeEntry.Name, models.FileExtension, ""))
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var startOid *git.Oid = branch.Target()

	revWalk, err := repo.Walk()
	revWalk.Sorting(git.SortTime)
	revWalk.Push(startOid)
	if err != nil {
		return nil, err
	}

	oID := startOid
	var lastCommit *git.Commit = nil
	attributionMap := map[string]string{}
	commitTimeMap := map[string]time.Time{}
	for {
		err := revWalk.Next(oID)
		if err != nil {
			break
		}

		commit, _ := repo.LookupCommit(oID)
		if lastCommit != nil {

			commitTree, err := commit.Tree()
			if err != nil {
				return nil, err
			}

			lastCommitTree, err := lastCommit.Tree()
			if err != nil {
				return nil, err
			}

			diff, err := repo.DiffTreeToTree(commitTree, lastCommitTree, &git.DiffOptions{})
			if err != nil {
				return nil, err
			}

			authorEmail := lastCommit.Author().Email
			commitTime := lastCommit.Author().When

			diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
				if delta.NewFile.Path == delta.OldFile.Path {
					blockID := strings.ReplaceAll(delta.NewFile.Path, models.FileExtension, "")
					if _, exists := attributionMap[blockID]; !exists {
						if utils.ExistsIn(blockID, blockIDs) {
							attributionMap[blockID] = authorEmail
							commitTimeMap[blockID] = commitTime
						}
					}
				}
				return nil, nil
			}, git.DiffDetailFiles)

			if len(blockIDs) == len(attributionMap) {
				break
			}

		}
		lastCommit = commit
	}

	if len(blockIDs) != len(attributionMap) {
		authorEmail := lastCommit.Author().Email
		commitTime := lastCommit.Author().When
		for _, blockID := range blockIDs {
			if _, exists := attributionMap[blockID]; !exists {
				attributionMap[blockID] = authorEmail
				commitTimeMap[blockID] = commitTime
			}
		}
	}

	attributions := []models.BlockAttribution{}
	for key, value := range attributionMap {
		attributions = append(attributions, models.BlockAttribution{
			AuthorEmail: value,
			BlockID:     key,
			UpdatedAt:   commitTimeMap[key],
		})
	}

	return &attributions, nil
}

func GetBranchAllAttributions(repo *git.Repository, branchName string, lastSyncedCommitID string) (*[]models.Attribution, string, error) {

	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return &[]models.Attribution{}, "", nil
	}

	var startOid *git.Oid = branch.Target()

	if startOid == nil {
		return &[]models.Attribution{}, "", nil
	}

	startCommitID := startOid.String()

	if startCommitID == lastSyncedCommitID {
		return &[]models.Attribution{}, startCommitID, nil
	}

	revWalk, err := repo.Walk()
	revWalk.Sorting(git.SortTime)
	revWalk.Push(startOid)
	if err != nil {
		return nil, "", err
	}

	syncInitialCommit := false

	oID := startOid
	var lastCommit *git.Commit = nil
	userEmailIds := map[string]int{}
	for {

		err := revWalk.Next(oID)
		if err != nil {
			if err.(*git.GitError).Code == git.ErrorCodeIterOver {
				syncInitialCommit = true
			}
			break
		}

		commit, _ := repo.LookupCommit(oID)
		if lastCommit != nil {
			commitTree, err := commit.Tree()
			if err != nil {
				return nil, "", err
			}

			lastCommitTree, err := lastCommit.Tree()
			if err != nil {
				return nil, "", err
			}

			diff, err := repo.DiffTreeToTree(commitTree, lastCommitTree, &git.DiffOptions{})
			if err != nil {
				return nil, "", err
			}

			authorEmail := lastCommit.Author().Email

			fmt.Println(diff.NumDeltas())

			diff.ForEach(func(delta git.DiffDelta, progress float64) (git.DiffForEachHunkCallback, error) {
				if authorEmail == "hey@bip.so" || authorEmail == "6ba19b04-0b96-421c-82ec-caf154110464@bip.so" || authorEmail == "52221613-8b61-4412-bcb6-1a0f26699764@bip.so" {
					return nil, nil
				} else if delta.NewFile.Path == delta.OldFile.Path && delta.NewFile.Path != models.RanksFileName {
					userEmailIds[authorEmail] = userEmailIds[authorEmail] + 1
				}
				return nil, nil
			}, git.DiffDetailFiles)
		}

		if lastSyncedCommitID == oID.String() {
			break
		}

		lastCommit = commit
	}

	if lastCommit != nil && syncInitialCommit {
		initialCommitAuthorEmail := lastCommit.Author().Email
		initialCommitTree, err := lastCommit.Tree()
		if err != nil {
			return nil, "", err
		}

		err = initialCommitTree.Walk(func(_ string, treeEntry *git.TreeEntry) error {
			if treeEntry.Name != models.RanksFileName {
				userEmailIds[initialCommitAuthorEmail] = userEmailIds[initialCommitAuthorEmail] + 1
				return nil
			}
			return nil
		})
		if err != nil {
			return nil, "", err
		}
	}

	attributions := []models.Attribution{}
	for key, value := range userEmailIds {
		attributions = append(attributions, models.Attribution{
			AuthorEmail: key,
			Edits:       value,
		})
	}

	return &attributions, startCommitID, nil
}
