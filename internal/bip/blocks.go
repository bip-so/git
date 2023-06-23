package bip

import (
	"bipgit/internal/constants"
	"bipgit/internal/libgit"
	"bipgit/internal/models"
	"errors"

	git "github.com/libgit2/git2go/v33"
)

func CommitNewMessageBlock(repo *git.Repository, sign *git.Signature, blocks models.BlocksSlice, branchName string, newMessageBlockID string) (string, error) {

	currentBlocks, err := libgit.ReadAllBlocksFromBranch(repo, sign, branchName)
	if err != nil {
		if err.Error() == "invalid branch" && branchName == constants.DefaultBranchName {
			currentBlocks = models.BlocksSlice{}
		} else {
			return "", err
		}
	}

	// TODO: Verify the logic if correct and is needed in case of ranks!!!
	allBelowBlocksPos := map[string]int32{}
	var newMessageBlock *models.BlockV2 = nil
	for _, blk := range blocks {
		if newMessageBlock != nil {
			allBelowBlocksPos[blk.ID] = blk.Rank
		}
		if blk.ID == newMessageBlockID {
			newMessageBlock = blk
		}
	}

	if newMessageBlock == nil {
		return "", errors.New("cannot find new message block in blocks")
	}

	var insertMessageBlkAt int32 = -1
	for _, blk := range currentBlocks {
		if _, exists := allBelowBlocksPos[blk.ID]; exists {
			insertMessageBlkAt = blk.Rank
			break
		}
	}
	if insertMessageBlkAt == -1 {
		insertMessageBlkAt = int32(len(currentBlocks))
	}

	for i, blk := range currentBlocks {
		if blk.Rank >= insertMessageBlkAt {
			currentBlocks[i].Rank += 1
		}
	}
	newMessageBlock.Rank = insertMessageBlkAt
	currentBlocks = append(currentBlocks, newMessageBlock)
	currentBlocks.Sort()

	return libgit.WriteBlocksAsCommit(repo, sign, currentBlocks, branchName, "Message Block Added")
}
