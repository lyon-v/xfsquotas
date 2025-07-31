/*
 * Copyright The moby Authors
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
 * This file is copied from https://github.com/moby/moby/blob/master/quota/projectquota.go
 * and chose the functions how to set/get project quota via system call
 */

package project

/*
#include <stdlib.h>
#include <dirent.h>
#include <linux/fs.h>
#include <linux/quota.h>
#include <linux/dqblk_xfs.h>

#ifndef FS_XFLAG_PROJINHERIT
struct fsxattr {
	__u32		fsx_xflags;
	__u32		fsx_extsize;
	__u32		fsx_nextents;
	__u32		fsx_projid;
	unsigned char	fsx_pad[12];
};
#define FS_XFLAG_PROJINHERIT	0x00000200
#endif
#ifndef FS_IOC_FSGETXATTR
#define FS_IOC_FSGETXATTR		_IOR ('X', 31, struct fsxattr)
#endif
#ifndef FS_IOC_FSSETXATTR
#define FS_IOC_FSSETXATTR		_IOW ('X', 32, struct fsxattr)
#endif

#ifndef PRJQUOTA
#define PRJQUOTA	2
#endif
#ifndef XFS_PROJ_QUOTA
#define XFS_PROJ_QUOTA	2
#endif
#ifndef Q_XSETPQLIM
#define Q_XSETPQLIM QCMD(Q_XSETQLIM, PRJQUOTA)
#endif
#ifndef Q_XGETPQUOTA
#define Q_XGETPQUOTA QCMD(Q_XGETQUOTA, PRJQUOTA)
#endif

const int Q_XGETQSTAT_PRJQUOTA = QCMD(Q_XGETQSTAT, PRJQUOTA);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"xfsquotas/internal/mount"

	"golang.org/x/sys/unix"
)

// DiskQuotaSize group disk quota size
type DiskQuotaSize struct {
	Quota      uint64 `json:"quota"`
	Inodes     uint64 `json:"inodes"`
	QuotaUsed  uint64 `json:"-"`
	InodesUsed uint64 `json:"-"`
}

const (
	noQuotaID    quotaID = 0
	firstQuotaID quotaID = 1048577
	// Don't go into an infinite loop searching for an unused quota id
	maxSearch        = 256
	idNameSeprator   = "-"
	quotaMountOption = "prjquota"
)
const (
	projIdNoCreate = true
	persistToFile  = true
)

var NotSupported = errors.New("not suppported")

// quotaID is generic quota identifier.
// Data type based on quotactl(2).
type quotaID int32

// String returns quota id in string format
func (q quotaID) String() string {
	return fmt.Sprintf("%d", q)
}

// IdName returns quota id with path flag
func (q quotaID) IdName(projectName string) string {
	return fmt.Sprintf("%s%s%d", projectName, idNameSeprator, q)
}

// ProjectQuota is the struct of project quota
type ProjectQuota struct {
	// path => device
	pathMapBackingDev map[string]*backingDev
	// id => id with path flag
	idNames map[quotaID]string
	// id => path
	idPaths map[quotaID][]string
	// path => id
	pathIds map[string]quotaID
	// store project quota to file
	nameIds map[string]quotaID
	prjFile *projectFile
}

type backingDev struct {
	supported bool
	device    string
}

// NewProjectQuota creates a new ProjectQuota
func NewProjectQuota() *ProjectQuota {
	return &ProjectQuota{
		pathMapBackingDev: make(map[string]*backingDev),
		idNames:           make(map[quotaID]string),
		idPaths:           make(map[quotaID][]string),
		pathIds:           make(map[string]quotaID),
		nameIds:           make(map[string]quotaID),
		prjFile:           NewProjectFile(),
	}
}

// GetQuota returns the quota for the given path
func (p *ProjectQuota) GetQuota(targetPath string) (*DiskQuotaSize, error) {
	backingDev, err := p.findAvailableBackingDev(targetPath)
	if err != nil {
		return nil, err
	}
	if !backingDev.supported {
		return nil, NotSupported
	}
	projectID, err := getProjectID(targetPath)
	if err != nil {
		return nil, err
	}
	return getProjectQuota(backingDev.device, projectID)
}

// SetQuota sets the quota for the given path
func (p *ProjectQuota) SetQuota(targetPath string, size *DiskQuotaSize) error {
	backingDev, err := p.findOrCreateBackingDev(targetPath)
	if err != nil {
		return err
	}
	if !backingDev.supported {
		return NotSupported
	}
	projectID, _, err := p.findOrCreateProjectId(targetPath, projIdNoCreate, persistToFile)
	if err != nil {
		return err
	}
	return setProjectQuota(backingDev.device, projectID, size)
}

// ClearQuota clears the quota for the given path
func (p *ProjectQuota) ClearQuota(targetPath string) error {
	backingDev, err := p.findAvailableBackingDev(targetPath)
	if err != nil {
		return err
	}
	if !backingDev.supported {
		return NotSupported
	}
	projectID, err := getProjectID(targetPath)
	if err != nil {
		return err
	}
	// Clear the quota
	return setProjectQuota(backingDev.device, projectID, &DiskQuotaSize{
		Quota:  0,
		Inodes: 0,
	})
}

// findAvailableBackingDev find available backing device for the path
func (p *ProjectQuota) findAvailableBackingDev(targetPath string) (*backingDev, error) {
	mount, err := mount.FindMount(targetPath)
	if err != nil {
		return nil, err
	}
	backingDev := &backingDev{
		supported: mount.FilesystemType == "xfs",
		device:    mount.Device,
	}
	return backingDev, nil
}

// findOrCreateBackingDev find or create backing device for the path
func (p *ProjectQuota) findOrCreateBackingDev(targetPath string) (*backingDev, error) {
	backingDev, err := p.findAvailableBackingDev(targetPath)
	if err != nil {
		return nil, err
	}
	if !backingDev.supported {
		return nil, NotSupported
	}
	return backingDev, nil
}

// findOrCreateSharedProjectId check if the path already has an shared project id, creating if not.
func (p *ProjectQuota) findOrCreateSharedProjectId(targetPath, projName string) (quotaID, bool, error) {
	isNewId := false
	projectID, exists := p.nameIds[projName]
	if !exists {
		projectID = p.allocateProjectID(targetPath)
		p.nameIds[projName] = projectID
		p.idNames[projectID] = projName
		isNewId = true
	}
	return projectID, isNewId, nil
}

// bindProjectId bind project id to the path
func (p *ProjectQuota) bindProjectId(targetPath string, projectId quotaID) (bool, error) {
	// Check if the path already has a project id
	existingProjectID, err := getProjectID(targetPath)
	if err != nil {
		return false, err
	}
	if existingProjectID == projectId {
		return false, nil
	}
	// Set the project id
	if err := setProjectID(targetPath, projectId); err != nil {
		return false, err
	}
	return true, nil
}

// findOrCreateProjectId find or create project id for the path
func (p *ProjectQuota) findOrCreateProjectId(targetPath string,
	noCreate bool, persist bool) (quotaID, bool, error) {
	isNewId := false
	projectID, exists := p.pathIds[targetPath]
	if !exists {
		if noCreate {
			return noQuotaID, false, fmt.Errorf("project id not found for path %s", targetPath)
		}
		projectID = p.allocateProjectID(targetPath)
		p.pathIds[targetPath] = projectID
		p.idPaths[projectID] = append(p.idPaths[projectID], targetPath)
		isNewId = true
	}
	return projectID, isNewId, nil
}

// allocateProjectID allocate a new project id
func (p *ProjectQuota) allocateProjectID(targetPath string) quotaID {
	// Simple allocation strategy: start from firstQuotaID and find an unused one
	for i := firstQuotaID; i < firstQuotaID+maxSearch; i++ {
		// Check if this ID is already used
		used := false
		for _, paths := range p.idPaths {
			for _, path := range paths {
				if path == targetPath {
					used = true
					break
				}
			}
			if used {
				break
			}
		}
		if !used {
			return i
		}
	}
	// If we can't find an unused ID, just return the next one
	return firstQuotaID + maxSearch
}

func getProjectQuota(backingFsBlockDev string, projectID quotaID) (*DiskQuotaSize, error) {
	var dqblk C.struct_fs_disk_quota
	var cs = C.CString(backingFsBlockDev)
	defer C.free(unsafe.Pointer(cs))

	_, _, errno := unix.Syscall6(unix.SYS_QUOTACTL, C.Q_XGETPQUOTA,
		uintptr(unsafe.Pointer(cs)), uintptr(projectID),
		uintptr(unsafe.Pointer(&dqblk)), 0, 0)
	if errno != 0 {
		return nil, fmt.Errorf("failed to get quota for project %d: %v", projectID, errno)
	}

	return &DiskQuotaSize{
		Quota:      uint64(dqblk.d_blk_hardlimit) * 512,
		Inodes:     uint64(dqblk.d_ino_hardlimit),
		QuotaUsed:  uint64(dqblk.d_bcount) * 512,
		InodesUsed: uint64(dqblk.d_icount),
	}, nil
}

func setProjectQuota(backingFsBlockDev string, projectID quotaID, quota *DiskQuotaSize) error {
	var dqblk C.struct_fs_disk_quota
	var cs = C.CString(backingFsBlockDev)
	defer C.free(unsafe.Pointer(cs))

	// Set the quota limits
	dqblk.d_blk_hardlimit = C.__u64(quota.Quota / 512)
	dqblk.d_ino_hardlimit = C.__u64(quota.Inodes)

	_, _, errno := unix.Syscall6(unix.SYS_QUOTACTL, C.Q_XSETPQLIM,
		uintptr(unsafe.Pointer(cs)), uintptr(projectID),
		uintptr(unsafe.Pointer(&dqblk)), 0, 0)
	if errno != 0 {
		return fmt.Errorf("failed to set quota for project %d: %v", projectID, errno)
	}

	return nil
}

func getProjectID(targetPath string) (quotaID, error) {
	var fsx C.struct_fsxattr
	var cs = C.CString(targetPath)
	defer C.free(unsafe.Pointer(cs))

	_, _, errno := unix.Syscall6(unix.SYS_IOCTL, 0,
		uintptr(unsafe.Pointer(cs)), uintptr(C.FS_IOC_FSGETXATTR),
		uintptr(unsafe.Pointer(&fsx)), 0, 0)
	if errno != 0 {
		return 0, fmt.Errorf("failed to get project id for %s: %v", targetPath, errno)
	}

	return quotaID(fsx.fsx_projid), nil
}

func setProjectID(targetPath string, projectID quotaID) error {
	var fsx C.struct_fsxattr
	var cs = C.CString(targetPath)
	defer C.free(unsafe.Pointer(cs))

	// Get current attributes
	_, _, errno := unix.Syscall6(unix.SYS_IOCTL, 0,
		uintptr(unsafe.Pointer(cs)), uintptr(C.FS_IOC_FSGETXATTR),
		uintptr(unsafe.Pointer(&fsx)), 0, 0)
	if errno != 0 {
		return fmt.Errorf("failed to get current attributes for %s: %v", targetPath, errno)
	}

	// Set the project ID
	fsx.fsx_projid = C.__u32(projectID)

	_, _, errno = unix.Syscall6(unix.SYS_IOCTL, 0,
		uintptr(unsafe.Pointer(cs)), uintptr(C.FS_IOC_FSSETXATTR),
		uintptr(unsafe.Pointer(&fsx)), 0, 0)
	if errno != 0 {
		return fmt.Errorf("failed to set project id for %s: %v", targetPath, errno)
	}

	return nil
}

func free(p *C.char) {
	C.free(unsafe.Pointer(p))
}

func openDir(path string) (*C.DIR, error) {
	Cpath := C.CString(path)
	defer free(Cpath)

	dir := C.opendir(Cpath)
	if dir == nil {
		return nil, fmt.Errorf("failed to open directory %s", path)
	}
	return dir, nil
}

func closeDir(dir *C.DIR) {
	C.closedir(dir)
}

func getDirFd(dir *C.DIR) uintptr {
	return uintptr(C.dirfd(dir))
}
