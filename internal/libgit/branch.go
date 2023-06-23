package libgit

import (
	"bipgit/internal/constants"
	"errors"

	git "github.com/libgit2/git2go/v33"
)

func CreateNewBranch(repo *git.Repository, fromBranchName, toBranchName string) (string, error) {

	toBranchExists := true
	toBranch, err := repo.LookupBranch(toBranchName, git.BranchLocal)
	if err != nil {
		gitErr := err.(*git.GitError)
		if gitErr.Code == git.ErrNotFound {
			toBranchExists = false
		} else {
			return "", err
		}
	}

	if toBranchExists {
		return toBranch.Target().String(), nil
	}

	if toBranchName != constants.DefaultBranchName {
		fromBranch, err := repo.LookupBranch(fromBranchName, git.BranchLocal)
		if err != nil {
			return "", err
		}

		commitTarget, err := repo.LookupCommit(fromBranch.Target())
		if err != nil {
			return "", err
		}

		toBranch, err := repo.CreateBranch(toBranchName, commitTarget, false)
		if err != nil {
			return "", err
		}
		return toBranch.Target().String(), err
	}
	return "", errors.New("trying to create " + constants.DefaultBranchName + " branch")
}

func CheckIfBranchExists(repo *git.Repository, branchName string) (bool, error) {
	branchExists := true
	_, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		gitErr := err.(*git.GitError)
		if gitErr.Code == git.ErrNotFound {
			branchExists = false
		} else {
			return false, err
		}
	}
	return branchExists, nil
}

func DeleteBranch(repo *git.Repository, branchName string) error {
	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return err
	}
	return branch.Delete()
}
