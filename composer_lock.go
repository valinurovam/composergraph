package composergraph

import "encoding/json"

type composerLock struct {
	Packages    []*composerJSON `json:"packages"`
	PackagesDev []*composerJSON `json:"packages-dev"`
}

func parseComposerLock(lockData []byte) (*composerLock, error) {
	lockFile := &composerLock{}
	if err := json.Unmarshal(lockData, lockFile); err != nil {
		return nil, err
	}

	return lockFile, nil
}
