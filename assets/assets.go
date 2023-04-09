// Copyright 2023 Deflinhec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/text/language"
)

//go:embed *
var embedFS embed.FS

type templateFS struct {
	tags []language.Tag
	fs   fs.FS
}

func (fs *templateFS) With(lang ...language.Tag) {
	fs.tags = lang
}

func (fs *templateFS) SetDataDir(dir string) {
	fs.fs = os.DirFS(filepath.Join(dir, "assets"))
}

func (fs *templateFS) Open(name string) (fs.File, error) {
	dir := filepath.Dir(name)
	base := filepath.Base(name)
	for _, tag := range fs.tags {
		path := filepath.Join(dir, tag.String(), base)
		if f, err := fs.fs.Open(path); err == nil {
			return f, nil
		} else if f, err := embedFS.Open(path); err == nil {
			return f, nil
		}
		b, _ := tag.Base()
		path = filepath.Join(dir, b.String(), base)
		if f, err := fs.fs.Open(path); err == nil {
			return f, nil
		} else if f, err := embedFS.Open(path); err == nil {
			return f, nil
		}
	}
	f, err := fs.fs.Open(name)
	if err != nil {
		return embedFS.Open(name)
	}
	return f, nil
}

var FS = &templateFS{fs: os.DirFS("assets")}
