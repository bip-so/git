package models

type ApiResponse struct{}

// Swagger models which are used to refernce in the swagger docs

type CreateSnapshotData struct {
	Blocks         BlocksSlice `json:"blocks"`
	BranchName     string      `json:"branchName"`
	FromBranchName string      `json:"fromBranchName"`
	Message        string      `json:"message"`
}

type CreateBlockSnapshotData struct {
	Blocks         BlocksSlice `json:"blocks"`
	BranchName     string      `json:"branchName"`
	MessageBlockID string      `json:"messageBlockId"`
}

type MergeBranchesData struct {
	ToBranchName              string                 `json:"toBranchName"`
	FromBranchName            string                 `json:"fromBranchName"`
	MergeType                 string                 `json:"mergeType"`
	FromBranchCreatedCommitID string                 `json:"fromBranchCreatedCommitId"`
	ChangesAccepted           map[string]interface{} `json:"changesAccepted"`
}
