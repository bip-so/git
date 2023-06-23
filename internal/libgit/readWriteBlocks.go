package libgit

import (
	"bipgit/internal/constants"
	"bipgit/internal/models"
	"errors"

	git "github.com/libgit2/git2go/v33"
)

func blocksFromCommitTree(commitTree *git.Tree, repo *git.Repository) (models.BlocksSlice, error) {
	blocks := models.BlocksSlice{}
	var positionsDataStr *string

	err := commitTree.Walk(func(_ string, treeEntry *git.TreeEntry) error {
		blob, err := repo.LookupBlob(treeEntry.Id)
		if err != nil {
			return git.MakeGitError2(int(git.ErrorCodeGeneric))
		}
		data := string(blob.Contents())
		if treeEntry.Name == models.RanksFileName {
			positionsDataStr = &data
			return nil
		}
		block, err := models.CreateV2BlockFromJson(data)
		if err != nil {
			return git.MakeGitError2(int(git.ErrorCodeGeneric))
		}
		block.SetFileName(treeEntry.Name)
		blocks = append(blocks, block)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if positionsDataStr == nil {
		return nil, errors.New("positionsDataStr: positions not found for blocks")
	}
	blocks.ReadRanksDotYaml(*positionsDataStr)
	return blocks, err
}

func WriteBlocksAsCommit(repo *git.Repository, signature *git.Signature, blocks models.BlocksSlice, branchName string, message string) (string, error) {
	var treeBuilder *git.TreeBuilder
	var isFirstCommit bool = false

	head, err := repo.Head()
	var commitTarget *git.Oid
	if err != nil {
		treeBuilder, err = repo.TreeBuilder()
		if err != nil {
			return "", err
		}
		isFirstCommit = true
	} else {
		commitTarget = head.Target()

		if branchName != constants.DefaultBranchName {
			branch, err := repo.LookupBranch(branchName, git.BranchLocal)
			if err != nil {
				return "", err
			}
			commitTarget = branch.Target()
		}

		commitTargetOld, err := repo.LookupCommit(commitTarget)
		if err != nil {
			return "", err
		}

		commitTreeOld, err := commitTargetOld.Tree()
		if err != nil {
			return "", err
		}

		treeBuilder, err = commitTreeOld.Owner().TreeBuilder()
		if err != nil {
			return "", err
		}
	}

	for _, block := range blocks {
		str, err := block.ToJson()
		if err != nil {
			return "", err
		}
		blobbytes := []byte(str)
		blobOid, err := repo.CreateBlobFromBuffer(blobbytes)
		if err != nil {
			return "", err
		}
		err = treeBuilder.Insert(block.FileName(), blobOid, git.FilemodeBlob)
		if err != nil {
			return "", err
		}
	}
	str, err := blocks.CreateRanksDotYaml()
	if err != nil {
		return "", err
	}
	blobbytes := []byte(str)
	blobOid, err := repo.CreateBlobFromBuffer(blobbytes)
	if err != nil {
		return "", err
	}
	err = treeBuilder.Insert(models.RanksFileName, blobOid, git.FilemodeBlob)
	if err != nil {
		return "", err
	}

	treeId, err := treeBuilder.Write()
	if err != nil {
		return "", err
	}

	var oID *git.Oid
	if isFirstCommit {
		oID, err = repo.CreateCommitFromIds("refs/heads/"+branchName, signature, signature, message, treeId)
	} else {
		oID, err = repo.CreateCommitFromIds("refs/heads/"+branchName, signature, signature, message, treeId, commitTarget)
	}
	if err != nil {
		return "", err
	}
	return oID.String(), nil
}

func ReadAllBlocksFromBranch(repo *git.Repository, signature *git.Signature, branchName string) (models.BlocksSlice, error) {

	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return nil, errors.New("invalid branch")
	}

	commitTarget, err := repo.LookupCommit(branch.Target())
	if err != nil {
		return nil, err
	}

	commitTree, err := commitTarget.Tree()
	if err != nil {
		return nil, err
	}

	return blocksFromCommitTree(commitTree, repo)
}

func ReadAllBlocksFromCommit(repo *git.Repository, signature *git.Signature, commitId string) (models.BlocksSlice, error) {
	oid, _ := git.NewOid(commitId)

	commitTarget, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, err
	}

	commitTree, err := commitTarget.Tree()
	if err != nil {
		return nil, err
	}

	return blocksFromCommitTree(commitTree, repo)
}
