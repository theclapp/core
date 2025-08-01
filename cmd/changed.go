// Copyright (c) 2023, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/base/logx"
	"cogentcore.org/core/cmd/config"
)

// Changed concurrently prints all of the repositories within this directory
// that have been changed and need to be updated in Git.
func Changed(c *config.Config) error { //types:add
	wg := sync.WaitGroup{}
	errs := []error{}
	fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if d.Name() != ".git" {
				return
			}
			dir := filepath.Dir(path)
			out, err := exec.Major().SetDir(dir).Output("git", "diff")
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting diff of %q: %w", dir, err))
				return
			}
			if out != "" { // if we have a diff, we have been changed
				fmt.Println(logx.CmdColor(dir))
				return
			}
			// if we don't have a diff, we also check to make sure we aren't ahead of the remote
			out, err = exec.Minor().SetDir(dir).Output("git", "status")
			if err != nil {
				errs = append(errs, fmt.Errorf("error getting status of %q: %w", dir, err))
				return
			}
			if strings.Contains(out, "Your branch is ahead") { // if we are ahead, we have been changed
				fmt.Println(logx.CmdColor(dir))
			}
		}()
		return nil
	})
	wg.Wait()
	fmt.Println("")
	return errors.Join(errs...)
}
