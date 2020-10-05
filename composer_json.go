package composergraph

import "encoding/json"

type composerJSON struct {
	Name       string            `json:"name"`
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
	Replace    map[string]string `json:"replace"`
	Config     *config           `json:"config"`

	// composer.lock fields
	Package string      `json:"package"`
	Version string      `json:"version"`
	Source  *sourceData `json:"source"`
}

type config struct {
	VendorDir string `json:"vendor-dir"`
}

type sourceData struct {
	Type      string `json:"type"`
	URL       string `json:"url"`
	Reference string `json:"reference"`
}

func parseComposerJSON(jsonData []byte) (*composerJSON, error) {
	jsonFile := &composerJSON{}
	if err := json.Unmarshal(jsonData, jsonFile); err != nil {
		return nil, err
	}

	return jsonFile, nil
}
