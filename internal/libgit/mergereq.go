package libgit

import (
	"bipgit/internal/constants"
	"bipgit/internal/models"
	"bipgit/internal/utils"
	"fmt"
	"sort"
	"strings"

	git "github.com/libgit2/git2go/v33"
	"gopkg.in/yaml.v2"
)

func MergeBranches(repo *git.Repository, signature *git.Signature, sourceBranchName string, destinationBranchName string, mergeType string, changesAccepted map[string]interface{}) (string, string, string, error) {

	sourceBranch, err := repo.LookupBranch(sourceBranchName, git.BranchLocal)
	if err != nil {
		return "", "", "", err
	}
	defer sourceBranch.Free()

	destinationBranch, err := repo.LookupBranch(destinationBranchName, git.BranchLocal)
	if err != nil {
		return "", "", "", err
	}
	defer destinationBranch.Free()

	ourTarget := destinationBranch.Target()
	ourTargetCommit, err := repo.LookupCommit(ourTarget)
	if err != nil {
		return "", "", "", err
	}

	theirTarget := sourceBranch.Target()
	theirTargetCommit, err := repo.LookupCommit(theirTarget)
	if err != nil {
		return "", "", "", err
	}

	mergeOpts, _ := git.DefaultMergeOptions()
	mergeOpts.FileFavor = git.MergeFileFavorTheirs

	index, err := repo.MergeCommits(ourTargetCommit, theirTargetCommit, &mergeOpts)
	if err != nil {
		return "", "", "", err
	}
	fmt.Println(index.EntryCount())
	for i := uint(0); i < index.EntryCount(); i++ {
		idxEntry, err := index.EntryByIndex(i)
		if err != nil {
			return "", "", "", err
		}
		fmt.Println(idxEntry.Path)
	}

	// resolve conflicts
	if index.HasConflicts() {
		conflicts, err := index.ConflictIterator()
		if err != nil {
			return "", "", "", err
		}
		defer conflicts.Free()
		for {
			c, err := conflicts.Next()
			if err != nil {
				gitErr := err.(*git.GitError)
				if gitErr.Code == git.ErrorCodeIterOver {
					break
				} else {
					return "", "", "", err
				}
			}
			if c.Our == nil && c.Their != nil {
				index.RemoveConflict(c.Their.Path)
				if err := index.Add(c.Their); err != nil {
					return "", "", "", fmt.Errorf("error resolving merge conflict for '%s': %v", c.Their.Path, err)
				}
			} else if c.Their == nil && c.Our != nil {
				index.RemoveConflict(c.Our.Path)
				if err := index.Add(c.Our); err != nil {
					return "", "", "", fmt.Errorf("error resolving merge conflict for '%s': %v", c.Our.Path, err)
				}
			} else if c.Their == nil && c.Our == nil {
				index.RemoveByPath(c.Ancestor.Path)
			}
		}
	}

	if mergeType == constants.MergeRequestPartiallyAccepted {
		err = partialMerge(index, ourTargetCommit, changesAccepted, repo)
		if err != nil {
			return "", "", "", err
		}
	}

	err = fixRanksFileAfterMerge(index, ourTargetCommit, theirTargetCommit, repo)
	if err != nil {
		return "", "", "", err
	}

	commit, err := repo.LookupCommit(sourceBranch.Target())
	if err != nil {
		return "", "", "", err
	}
	defer commit.Free()

	requestSignature := commit.Author()
	requestCommitMessage := "\"" + commit.Message() + "\" "
	if commit.Message() == constants.CommitMessageCreatingMergeRequest {
		requestCommitMessage = ""
	}

	treeId, err := index.WriteTreeTo(repo)
	if err != nil {
		return "", "", "", err
	}

	tree, err := repo.LookupTree(treeId)
	if err != nil {
		return "", "", "", err
	}
	defer tree.Free()

	message := fmt.Sprintf("Accepted %sMR from %s", requestCommitMessage, requestSignature.Name)

	oID, err := repo.CreateCommit("refs/heads/"+destinationBranchName, signature, signature, message,
		tree, ourTargetCommit, commit)
	if err != nil {

		return "", "", "", err
	}

	err = repo.StateCleanup()
	if err != nil {
		return "", "", "", err
	}

	return oID.String(), theirTarget.String(), ourTarget.String(), nil
}

func fixRanksFileAfterMerge(index *git.Index, ourTargetCommit *git.Commit, theirTargetCommit *git.Commit, repo *git.Repository) error {

	allblockIDs := []string{}
	var allRanks map[string]int32
	var ourRanks map[string]int32
	var ranksIdxEntry *git.IndexEntry

	ourCommitTree, err := ourTargetCommit.Tree()
	if err != nil {
		return err
	}
	ourPositionsTreeEntry := ourCommitTree.EntryByName(models.RanksFileName)
	blob, err := repo.LookupBlob(ourPositionsTreeEntry.Id)
	if err != nil {
		return err
	}
	ourPositionsDataStr := string(blob.Contents())
	err = yaml.Unmarshal([]byte(ourPositionsDataStr), &ourRanks)
	if err != nil {
		return err
	}

	theirCommitTree, err := theirTargetCommit.Tree()
	if err != nil {
		return err
	}
	theirPositionsTreeEntry := theirCommitTree.EntryByName(models.RanksFileName)
	blob, err = repo.LookupBlob(theirPositionsTreeEntry.Id)
	if err != nil {
		return err
	}
	theirPositionsDataStr := string(blob.Contents())
	err = yaml.Unmarshal([]byte(theirPositionsDataStr), &allRanks)
	if err != nil {
		return err
	}

	for i := uint(0); i < index.EntryCount(); i++ {
		idxEntry, err := index.EntryByIndex(i)
		if err != nil {
			return err
		}
		if idxEntry.Path == models.RanksFileName {
			ranksIdxEntry = idxEntry
		} else {
			allblockIDs = append(allblockIDs, strings.ReplaceAll(idxEntry.Path, models.NewFileExtension, ""))
		}
	}

	sort.Slice(allblockIDs, func(i, j int) bool {
		x, y := allblockIDs[i], allblockIDs[j]
		var xpos, ypos int32
		if pos, exists := allRanks[x]; exists {
			xpos = pos
		} else {
			xpos = allRanks[x]
		}
		if pos, exists := allRanks[y]; exists {
			ypos = pos
		} else {
			ypos = allRanks[y]
		}
		return xpos < ypos
	})

	for _, key := range allblockIDs {
		position, exists := allRanks[key]
		if !exists {
			oldPos, exists := ourRanks[key]
			if !exists {
				allRanks[key] = 0
			} else {
				var beforeKey string = ""
				var decrement int32 = 1
				for beforeKey == "" && decrement <= oldPos {
					for key, value := range ourRanks {
						if value == oldPos-decrement {
							beforeKey = key
							break
						}
					}
					if pos, exists := allRanks[beforeKey]; exists {
						position = pos + 1
					} else {
						beforeKey = ""
						decrement = decrement + 1
					}
				}
				if beforeKey == "" {
					var afterKey string = ""
					var increment int32 = 1
					for afterKey == "" && int(increment+oldPos) < len(ourRanks) {
						for key, value := range ourRanks {
							if value == oldPos+increment {
								afterKey = key
								break
							}
						}
						if pos, exists := allRanks[afterKey]; exists {
							position = pos
						} else {
							afterKey = ""
							increment = increment + 1
						}
					}
				}
				for key, pos := range allRanks {
					if pos >= position {
						allRanks[key] = pos + 1
					}
				}
				allRanks[key] = position
			}
		}
	}

	for key := range allRanks {
		if !utils.ExistsIn(key, allblockIDs) {
			delete(allRanks, key)
		}
	}

	keys := make([]string, 0, len(allRanks))
	for key := range allRanks {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return allRanks[keys[i]] < allRanks[keys[j]] })
	for i, key := range keys {
		allRanks[key] = int32((i + 1) * 1000)
	}

	data, err := yaml.Marshal(&allRanks)
	if err != nil {
		return err
	}

	// TODO: Do this only when positions are not in sync,
	// currently it creates blob even if the positions are in sync.
	blobOid, err := repo.CreateBlobFromBuffer(data)
	if err != nil {
		return err
	}
	ranksIdxEntry.Id = blobOid
	err = index.Add(ranksIdxEntry)
	if err != nil {
		return err
	}
	return nil
}

func partialMerge(index *git.Index, ourTargetCommit *git.Commit, changesAccepted map[string]interface{}, repo *git.Repository) error {

	if len(changesAccepted) == 0 {
		return nil
	}

	ourCommitTree, err := ourTargetCommit.Tree()
	if err != nil {
		return err
	}

	for key, value := range changesAccepted {
		idxEntry, err := index.EntryByPath(key+models.NewFileExtension, 0)
		if err != nil {
			continue
		}
		if isAccepted, isOK := value.(bool); isOK {
			if !isAccepted {
				ourTreeEntry := ourCommitTree.EntryByName(idxEntry.Path)
				if ourTreeEntry == nil {
					err = index.RemoveByPath(idxEntry.Path)
					if err != nil {
						return err
					}
				} else {
					idxEntry.Id = ourTreeEntry.Id
					err = index.Add(idxEntry)
					if err != nil {
						return err
					}
				}
			}
		} else if editedBlock, isOK := value.(map[string]interface{}); isOK {
			ourTreeEntry := ourCommitTree.EntryByName(idxEntry.Path)
			blob, err := repo.LookupBlob(ourTreeEntry.Id)
			if err != nil {
				return err
			}
			existingBlock, err := models.CreateV2BlockFromJson(string(blob.Contents()))
			if err != nil {
				return err
			}
			existingBlock.Children = editedBlock["data"].([]map[string]interface{})
			dataStr, err := existingBlock.ToJson()
			if err != nil {
				return err
			}
			blobOid, err := repo.CreateBlobFromBuffer([]byte(dataStr))
			if err != nil {
				return err
			}
			idxEntry.Id = blobOid
			err = index.Add(idxEntry)
			if err != nil {
				return err
			}
		}
		changesAccepted[key] = nil
	}

	// for i := uint(0); i < index.EntryCount(); i++ {
	// 	fmt.Println(i)
	// 	idxEntry, err := index.EntryByIndex(i)
	// 	index.EntryByPath()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	idxId := strings.ReplaceAll(idxEntry.Path, models.NewFileExtension, "")
	// 	fmt.Println(idxEntry.Path, i, index.EntryCount())

	// 	if value, exists := changesAccepted[idxId]; exists {
	// 		if isAccepted, isOK := value.(bool); isOK {
	// 			if !isAccepted {
	// 				ourTreeEntry := ourCommitTree.EntryByName(idxEntry.Path)
	// 				if ourTreeEntry == nil {
	// 					err = index.RemoveByPath(idxEntry.Path)
	// 					if err != nil {
	// 						return err
	// 					}
	// 				} else {
	// 					idxEntry.Id = ourTreeEntry.Id
	// 					err = index.Add(idxEntry)
	// 					if err != nil {
	// 						return err
	// 					}
	// 				}
	// 			}
	// 		} else if editedBlock, isOK := value.(map[string]interface{}); isOK {
	// 			ourTreeEntry := ourCommitTree.EntryByName(idxEntry.Path)
	// 			blob, err := repo.LookupBlob(ourTreeEntry.Id)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			existingBlock, err := models.CreateV2BlockFromJson(string(blob.Contents()))
	// 			if err != nil {
	// 				return err
	// 			}
	// 			existingBlock.Children = editedBlock["data"].([]map[string]interface{})
	// 			dataStr, err := existingBlock.ToJson()
	// 			if err != nil {
	// 				return err
	// 			}
	// 			blobOid, err := repo.CreateBlobFromBuffer([]byte(dataStr))
	// 			if err != nil {
	// 				return err
	// 			}
	// 			idxEntry.Id = blobOid
	// 			err = index.Add(idxEntry)
	// 			if err != nil {
	// 				return err
	// 			}
	// 		}
	// 		changesAccepted[idxId] = nil
	// 	}
	// }

	for key, value := range changesAccepted {
		if value == nil {
			continue
		}
		if isAccepted, isOK := value.(bool); isOK {
			if !isAccepted {
				ourTreeEntry := ourCommitTree.EntryByName(key + models.NewFileExtension)
				// FOR BACKWARD COMPATIBILITY WITH YAML FILES ?? NOT NEEDED IF ALL ARE CONVERTED IN JSON AT SEEDER!
				// if ourTreeEntry == nil {
				// 	ourTreeEntry = ourCommitTree.EntryByName(key + models.FileExtension)
				// }
				blob, err := repo.LookupBlob(ourTreeEntry.Id)
				if err != nil {
					return err
				}
				var idxEntry git.IndexEntry = git.IndexEntry{
					Id:   blob.Id(),
					Mode: git.FilemodeBlob,
					Path: key + models.NewFileExtension,
				}
				err = index.Add(&idxEntry)
				if err != nil {
					return err
				}
				changesAccepted[key] = nil
			}
		}
	}

	return nil
}
