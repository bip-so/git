package models

import (
	"encoding/json"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type BlockV2 struct {
	ID         string                   `json:"-"` //has to be uuid
	Version    uint                     `json:"version"`
	Type       string                   `json:"type"`
	Rank       int32                    `json:"-"`
	Children   []map[string]interface{} `json:"children"`
	Attributes map[string]interface{}   `json:"attributes"`
}

type BEBlock struct {
	UUID       string                   `json:"uuid"`
	Version    uint                     `json:"version"`
	Type       string                   `json:"type"`
	Rank       int32                    `json:"rank"`
	Children   []map[string]interface{} `json:"children"`
	Attributes map[string]interface{}   `json:"attributes"`
}

const (
	NewFileExtension = ".json"
	RanksFileName    = "ranks.yaml"
)

func CreateV2BlockFromJson(dataStr string) (*BlockV2, error) {
	var block BlockV2
	err := json.Unmarshal([]byte(dataStr), &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (block BlockV2) ToJson() (string, error) {
	data, err := json.Marshal(&block)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (block BlockV2) FileName() string {
	return block.ID + NewFileExtension
}

func (block *BlockV2) SetFileName(fileName string) {
	block.ID = strings.ReplaceAll(fileName, NewFileExtension, "")
}

type BlocksSlice []*BlockV2
type BEBlocksSlice []*BEBlock

func (blocks BlocksSlice) ConvertToBEBlocksFromBlocks() BEBlocksSlice {
	newBlocks := BEBlocksSlice{}
	for _, blk := range blocks {
		newBlocks = append(newBlocks, &BEBlock{
			UUID:       blk.ID,
			Version:    blk.Version,
			Type:       blk.Type,
			Rank:       blk.Rank,
			Attributes: blk.Attributes,
			Children:   blk.Children,
		})
	}
	return newBlocks
}

func (blocks BEBlocksSlice) ConvertToBlocksFromBEBlocks() []*BlockV2 {
	newBlocks := []*BlockV2{}
	for _, blk := range blocks {
		newBlocks = append(newBlocks, &BlockV2{
			ID:         blk.UUID,
			Version:    blk.Version,
			Type:       blk.Type,
			Rank:       blk.Rank,
			Attributes: blk.Attributes,
			Children:   blk.Children,
		})
	}
	return newBlocks
}

func (blocks BlocksSlice) Sort() {
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Rank < blocks[j].Rank
	})
}

func (blocks BlocksSlice) CreateRanksDotYaml() (string, error) {
	blocks.Sort()
	allRanks := map[string]int32{}
	for _, b := range blocks {
		allRanks[b.ID] = b.Rank
	}
	data, err := yaml.Marshal(&allRanks)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (blocks BlocksSlice) ReadRanksDotYaml(dataStr string) error {
	var allRanks map[string]int32
	err := yaml.Unmarshal([]byte(dataStr), &allRanks)
	if err != nil {
		return err
	}
	for _, block := range blocks {
		block.Rank = allRanks[block.ID]
	}
	blocks.Sort()
	return nil
}
