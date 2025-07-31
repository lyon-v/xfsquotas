package xfs

import (
	"fmt"
	"os/exec"
	"strings"
)

// Client provides an interface for XFS quota operations
type Client struct{}

// NewClient creates a new XFS client
func NewClient() *Client {
	return &Client{}
}

// GetQuota gets quota information using xfs_quota command
func (c *Client) GetQuota(path string) (string, error) {
	cmd := exec.Command("xfs_quota", "-x", "-c", fmt.Sprintf("report -p %s", path))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get quota: %v", err)
	}
	return string(output), nil
}

// SetQuota sets quota using xfs_quota command
func (c *Client) SetQuota(path, size, inodes string) error {
	cmd := exec.Command("xfs_quota", "-x", "-c", 
		fmt.Sprintf("limit -p bsoft=%s bhard=%s isoft=%s ihard=%s %s", 
			size, size, inodes, inodes, path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set quota: %v, output: %s", err, string(output))
	}
	return nil
}

// ClearQuota clears quota using xfs_quota command
func (c *Client) ClearQuota(path string) error {
	cmd := exec.Command("xfs_quota", "-x", "-c", 
		fmt.Sprintf("limit -p bsoft=0 bhard=0 isoft=0 ihard=0 %s", path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clear quota: %v, output: %s", err, string(output))
	}
	return nil
}

// IsXFSFilesystem checks if the given path is on an XFS filesystem
func (c *Client) IsXFSFilesystem(path string) (bool, error) {
	cmd := exec.Command("df", "-T", path)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check filesystem type: %v", err)
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return false, fmt.Errorf("unexpected df output format")
	}
	
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return false, fmt.Errorf("unexpected df output format")
	}
	
	return strings.Contains(fields[1], "xfs"), nil
} 