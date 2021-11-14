// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crane

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

// asterisk The character which is treated like a glob
const asterisk = "*"

type ExtractArgs map[string]string

// Extract writes the filesystem contents (as a tarball) of img to the specified directory. Args take a map of output directory
// and pattern to match the files to extract.
func Extract(img v1.Image, args ExtractArgs) error {
	fs := mutate.Extract(img)
	defer fs.Close()

	tr := tar.NewReader(fs)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		for dir, pattern := range args {
			absPath, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			if ok, _ := match(pattern, hdr.Name); ok {
				fi := hdr.FileInfo()
				fileName := hdr.Name
				absFileName := filepath.Join(absPath, fileName)
				if fi.Mode().IsDir() {
					if err := os.MkdirAll(absFileName, 0755); err != nil {
						return err
					}
					continue
				}

				fileDir := filepath.Dir(absFileName)
				_, err := os.Stat(fileDir)
				if os.IsNotExist(err) {
					if err := os.MkdirAll(fileDir, 0755); err != nil {
						return err
					}
				}

				// create new file with original file mode
				file, err := os.OpenFile(
					absFileName,
					os.O_RDWR|os.O_CREATE|os.O_TRUNC,
					0755,
				)
				if err != nil {
					return err
				}
				if _, err := io.Copy(file, tr); err != nil {
					return err
				}
				if err := file.Close(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// match takes a pattern with a glob (asterisk) and determines if the supplied path matches.
func match(pattern, path string) (bool, error) {
	if pattern == "" {
		return path == pattern, nil
	}

	if len(strings.Split(pattern, asterisk)) == 1 {
		return path == pattern, nil
	}

	if pattern == asterisk {
		return true, nil
	}

	return scanChunks(pattern, path)
}

// scanChunks loops through the provided path determining if it
func scanChunks(pattern, path string) (bool, error) {
	chunks := strings.Split(pattern, asterisk)

	for i := 0; i < len(chunks)-1; i++ {
		ix := strings.Index(path, chunks[i])
		if i == 0 {
			if !strings.HasPrefix(pattern, asterisk) && ix != 0 {
				return false, nil
			}
		} else {
			if ix < 0 {
				return false, nil
			}
		}
		path = path[ix+len(chunks[i]):]
	}

	return strings.HasSuffix(pattern, asterisk) || strings.HasSuffix(path, chunks[len(chunks)-1]), nil
}
