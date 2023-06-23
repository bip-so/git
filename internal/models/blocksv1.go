package models

import (
	"encoding/json"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type BlockV1 struct {
	ID                string   `yaml:"-" json:"id"`
	Text              string   `yaml:"text" json:"text"`
	Type              string   `yaml:"type" json:"type"`
	Position          int      `yaml:"-" json:"position"`
	UserID            string   `yaml:"userId" json:"userId"`
	TweetID           *string  `yaml:"tweetId" json:"tweetId"`
	URL               *string  `yaml:"url" json:"url"`
	Properties        string   `yaml:"properties" json:"properties"`
	MentionedUserIDs  []string `yaml:"mentionedUserIds" json:"mentionedUserIds"`
	MentionedGroupIDs []string `yaml:"mentionedGroupIds" json:"mentionedGroupIds"`
	MentionedPageIDs  []string `yaml:"mentionedPageIds" json:"mentionedPageIds"`
}

const (
	FileExtension = ".yaml"
	// TODO: NEEDED FOR BACKWARD COMPATIBILITY
	// PositionsFileName = "positions.yaml"
)

func CreateBlockV1FromYaml(dataStr string) (*BlockV1, error) {
	var block BlockV1
	err := yaml.Unmarshal([]byte(dataStr), &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func CreateBlockV1FromJson(dataStr string) (*BlockV1, error) {
	var block BlockV1
	err := json.Unmarshal([]byte(dataStr), &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (block BlockV1) ToYaml() (string, error) {
	data, err := yaml.Marshal(&block)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (block BlockV1) FileName() string {
	return block.ID + FileExtension
}

func (block *BlockV1) SetFileName(fileName string) {
	block.ID = strings.ReplaceAll(fileName, FileExtension, "")
}

type BlocksV1Slice []*BlockV1

func (blocks BlocksV1Slice) Sort() {
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Position < blocks[j].Position
	})
}

func (blocks BlocksV1Slice) CreatePositionsDotYaml() (string, error) {
	blocks.Sort()
	allPositions := map[string]int{}
	for _, b := range blocks {
		allPositions[b.ID] = b.Position
	}
	data, err := yaml.Marshal(&allPositions)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (blocks BlocksV1Slice) ReadPositionsDotYaml(dataStr string) error {
	var allPositions map[string]int
	err := yaml.Unmarshal([]byte(dataStr), &allPositions)
	if err != nil {
		return err
	}
	for _, block := range blocks {
		block.Position = allPositions[block.ID]
	}
	blocks.Sort()
	return nil
}
