package api

import (
	"strconv"

	"xfsquotas/internal/project"

	"github.com/docker/go-units"
)

// QuotaManager provides a high-level interface for managing XFS quotas
type QuotaManager struct {
	quota *project.ProjectQuota
}

// NewQuotaManager creates a new QuotaManager instance
func NewQuotaManager() *QuotaManager {
	return &QuotaManager{
		quota: project.NewProjectQuota(),
	}
}

// GetQuota returns the quota information for the given path
func (q *QuotaManager) GetQuota(path string) (*project.DiskQuotaSize, error) {
	return q.quota.GetQuota(path)
}

// SetQuota sets the quota for the given path
func (q *QuotaManager) SetQuota(path string, sizeVal, inodeVal string) error {
	size, err := units.RAMInBytes(sizeVal)
	if err != nil {
		return err
	}
	inodes, err := strconv.ParseUint(inodeVal, 10, 64)
	if err != nil {
		return err
	}
	return q.quota.SetQuota(path, &project.DiskQuotaSize{
		Quota:  uint64(size),
		Inodes: inodes,
	})
}

// CleanQuota clears the quota for the given path
func (q *QuotaManager) CleanQuota(path string) error {
	return q.quota.ClearQuota(path)
}
