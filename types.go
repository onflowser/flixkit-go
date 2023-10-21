package flixkit

import (
	"encoding/hex"
	"fmt"
	"regexp"

	"golang.org/x/crypto/sha3"
)

type Network struct {
	Address        string `json:"address"`
	FqAddress      string `json:"fq_address"`
	Contract       string `json:"contract"`
	Pin            string `json:"pin"`
	PinBlockHeight uint64 `json:"pin_block_height"`
}

type Argument struct {
	Index    int      `json:"index"`
	Type     string   `json:"type"`
	Messages Messages `json:"messages"`
	Balance  string   `json:"balance"`
}

type Title struct {
	I18N map[string]string `json:"i18n"`
}

type Description struct {
	I18N map[string]string `json:"i18n"`
}

type Messages struct {
	Title       *Title       `json:"title,omitempty"`
	Description *Description `json:"description,omitempty"`
}

type Dependencies map[string]Contracts
type Contracts map[string]Networks
type Networks map[string]Network
type Arguments map[string]Argument

type Data struct {
	Type         string       `json:"type"`
	Interface    string       `json:"interface"`
	Messages     Messages     `json:"messages"`
	Cadence      string       `json:"cadence"`
	Dependencies Dependencies `json:"dependencies"`
	Arguments    Arguments    `json:"arguments"`
}

type FlowInteractionTemplate struct {
	FType    string `json:"f_type"`
	FVersion string `json:"f_version"`
	ID       string `json:"id"`
	Data     Data   `json:"data"`
}

func (t *FlowInteractionTemplate) IsScript() bool {
	return t.Data.Type == "script"
}

func (t *FlowInteractionTemplate) IsTransaction() bool {
	return t.Data.Type == "transaction"
}

func (t *FlowInteractionTemplate) GetAndReplaceCadenceImports(networkName string) (string, error) {
	cadence := t.Data.Cadence

	for dependencyAddress, contracts := range t.Data.Dependencies {
		for contractName, networks := range contracts {
			network, ok := networks[networkName]
			if !ok {
				return "", fmt.Errorf("network %s not found for contract %s", networkName, contractName)
			}

			pattern := fmt.Sprintf(`import\s*%s\s*from\s*%s`, contractName, dependencyAddress)
			re, err := regexp.Compile(pattern)
			if err != nil {
				return "", fmt.Errorf("invalid regex pattern: %v", err)
			}

			replacement := fmt.Sprintf("import %s from %s", contractName, network.Address)
			cadence = re.ReplaceAllString(cadence, replacement)
		}
	}

	return cadence, nil
}

func (t *FlowInteractionTemplate) GetDescription() string {
	s := ""
	if t.Data.Messages.Description != nil &&
		t.Data.Messages.Description.I18N != nil {

		// relying on en-US for now, future we need to know what language to use
		value, exists := t.Data.Messages.Description.I18N["en-US"]
		if exists {
			s = value
		}
	}
	return s
}

func (msgs *Messages) GetTitleValue(placeholder string) string {
	s := placeholder
	if msgs.Title != nil &&
		msgs.Title.I18N != nil {
		// relying on en-US for now, future we need to know what language to use
		value, exists := msgs.Title.I18N["en-US"]
		if exists {
			s = value
		}
	}
	return s
}

func GenHash(utf8String string) string {
	hasher := sha3.New256()          // Create a new SHA3 256 hasher
	hasher.Write([]byte(utf8String)) // Write the utf8 string to the hasher
	hash := hasher.Sum(nil)          // Get the hash result

	return hex.EncodeToString(hash) // Convert the hash result to hex
}
