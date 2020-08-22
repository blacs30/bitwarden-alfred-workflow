package alfred

import (
	"fmt"
	aw "github.com/deanishe/awgo"
)

const tokenKey = "token"

func GetToken(wf *aw.Workflow) (string, error) {
	var err error

	token, err := wf.Keychain.Get(tokenKey)
	if err != nil {
		return "", fmt.Errorf("token not found in the keychain")
	}

	return token, nil
}

func SetToken(wf *aw.Workflow, token string) error {
	return wf.Keychain.Set(tokenKey, token)
}

func RemoveToken(wf *aw.Workflow) error {
	return wf.Keychain.Delete(tokenKey)
}
