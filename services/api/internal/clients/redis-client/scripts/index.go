package scripts

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
)

type Script struct {
	Name string
	Src  string
	SHA  string
}

func New(name string) (*Script, error) {
	_, currentFile, _, _ := runtime.Caller(0)

	currentDir := filepath.Dir(currentFile)

	path := filepath.Join(currentDir, name)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	src := string(content)

	hash := sha1.Sum(content)

	return &Script{
		Name: name,
		Src:  src,
		SHA:  hex.EncodeToString(hash[:]),
	}, nil
}
