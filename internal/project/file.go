/*
 * Copyright The Kubernetes Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this file except in compliance with the License. You may obtain a copy of
 * the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 *
 * For file operations, we learn something from
 * https://github.com/kubernetes/kubernetes/pkg/volume/util/fsquota/project.go
 */

package project

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

var (
	tmpPrefix = "."

	projectsPath = "/etc/projects"
	projidPath   = "/etc/projid"
)

// projectFile used to record project quota to file
type projectFile struct{}

// NewProjectFile new project file instance
func NewProjectFile() *projectFile {
	//if !types.InHostNamespace {
	//	projectsPath = path.Join(types.RootFS, projectsPath)
	//	projidPath = path.Join(types.RootFS, projidPath)
	//}
	if err := projFilesAreOK(); err != nil {
		klog.Fatalf("project files are not ok: %v", err)
	}

	return &projectFile{}
}

// DumpProjectIds read project quota record
func (f *projectFile) DumpProjectIds() (idPaths map[quotaID][]string, idNames map[quotaID]string, err error) {
	idPaths = make(map[quotaID][]string)
	idNames = make(map[quotaID]string)

	// 1048579:/data1/test
	idPathHandleFunc := func(ranges []string) {
		idStr := strings.TrimSpace(ranges[0])
		id, err := strconv.Atoi(idStr)
		if err != nil {
			klog.Errorf("invalid project id(%s): %v", idStr, err)
			return
		}

		paths := []string{}
		pathName := strings.TrimSpace(ranges[1])
		if loadedPaths, exist := idPaths[quotaID(id)]; exist == true {
			paths = append(loadedPaths, pathName)
		} else {
			paths = append(paths, pathName)
		}

		idPaths[quotaID(id)] = paths
	}
	// emptyDir-1048577:1048577
	idNameHandleFunc := func(ranges []string) {
		idName := strings.TrimSpace(ranges[0])
		idStr := strings.TrimSpace(ranges[1])
		id, err := strconv.Atoi(idStr)
		if err != nil {
			klog.Errorf("invalid project id(%s): %v", idStr, err)
			return
		}
		idNames[quotaID(id)] = idName
	}
	dumpProjectsFile(projectsPath, idPathHandleFunc)
	dumpProjectsFile(projidPath, idNameHandleFunc)

	klog.V(2).Infof("dump new project paths: %+v", idPaths)
	klog.V(2).Infof("dump new project ids: %+v", idNames)
	return idPaths, idNames, nil
}

// UpdateProjects save projectid:path to /etc/projects
func (f *projectFile) UpdateProjects(idPaths map[quotaID][]string) error {
	content := ""
	for id, paths := range idPaths {
		for _, path := range paths {
			content += fmt.Sprintf("%d:%s\n", id, path)
		}
	}
	return writeByTempFile(projectsPath, []byte(content))
}

// UpdateProjIds save projectid:name to /etc/projid
func (f *projectFile) UpdateProjIds(idNames map[quotaID]string) error {
	content := ""
	for id, name := range idNames {
		content += fmt.Sprintf("%s:%d\n", name, id)
	}
	return writeByTempFile(projidPath, []byte(content))
}

func projFilesAreOK() error {
	// check if the project files exist and are writable
	if _, err := os.Stat(projectsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(projectsPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %v", projectsPath, err)
		}
		if _, err := os.Create(projectsPath); err != nil {
			return fmt.Errorf("failed to create %s: %v", projectsPath, err)
		}
	}
	if _, err := os.Stat(projidPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(projidPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %v", projidPath, err)
		}
		if _, err := os.Create(projidPath); err != nil {
			return fmt.Errorf("failed to create %s: %v", projidPath, err)
		}
	}
	return nil
}

func dumpProjectsFile(filePath string, handle func(ranges []string)) {
	file, err := os.Open(filePath)
	if err != nil {
		klog.Errorf("failed to open %s: %v", filePath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ranges := strings.Split(line, ":")
		if len(ranges) != 2 {
			klog.Errorf("invalid line in %s: %s", filePath, line)
			continue
		}
		handle(ranges)
	}
}

func writeByTempFile(pathFile string, data []byte) (retErr error) {
	dir := filepath.Dir(pathFile)
	tmpFile, err := ioutil.TempFile(dir, tmpPrefix)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer func() {
		if retErr != nil {
			os.Remove(tmpFile.Name())
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	if err := os.Rename(tmpFile.Name(), pathFile); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %v", pathFile, err)
	}
	return nil
}
