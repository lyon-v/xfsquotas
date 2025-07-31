package test

import (
	"testing"

	"xfsquotas/internal/project"
)

func TestDiskQuotaSize(t *testing.T) {
	quota := &project.DiskQuotaSize{
		Quota:      1024,
		Inodes:     100,
		QuotaUsed:  512,
		InodesUsed: 50,
	}

	if quota.Quota != 1024 {
		t.Errorf("Expected Quota to be 1024, got %d", quota.Quota)
	}

	if quota.Inodes != 100 {
		t.Errorf("Expected Inodes to be 100, got %d", quota.Inodes)
	}

	if quota.QuotaUsed != 512 {
		t.Errorf("Expected QuotaUsed to be 512, got %d", quota.QuotaUsed)
	}

	if quota.InodesUsed != 50 {
		t.Errorf("Expected InodesUsed to be 50, got %d", quota.InodesUsed)
	}
}

func TestNewProjectQuota(t *testing.T) {
	pq := project.NewProjectQuota()
	if pq == nil {
		t.Error("Expected NewProjectQuota to return non-nil")
	}
}
