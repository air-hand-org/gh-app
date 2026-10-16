package dependency

import (
	"os/exec"

	"github.com/cockroachdb/errors"
)

type Result struct {
	Name string
	Path string
	Err  error
}

func Check(names ...string) ([]Result, error) {
	results := make([]Result, 0, len(names))
	var checkErr error
	for _, name := range names {
		path, err := exec.LookPath(name)
		results = append(results, Result{Name: name, Path: path, Err: err})
		checkErr = errors.Join(checkErr, err)
	}
	return results, checkErr
}
