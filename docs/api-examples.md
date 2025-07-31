# API 使用示例

本文档提供了详细的 API 使用示例，展示如何在不同场景下使用 XFS Quota Manager。

## 基本用法

### 1. 简单的配额管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

func main() {
    // 创建配额管理器
    quota := api.NewQuotaManager()
    
    // 设置配额
    err := quota.SetQuota("/data/user1", "10GiB", "1000000")
    if err != nil {
        log.Fatal("设置配额失败:", err)
    }
    fmt.Println("配额设置成功")
    
    // 查询配额
    quotaInfo, err := quota.GetQuota("/data/user1")
    if err != nil {
        log.Fatal("查询配额失败:", err)
    }
    
    fmt.Printf("配额大小: %d bytes (%.2f GB)\n", 
        quotaInfo.Quota, float64(quotaInfo.Quota)/1024/1024/1024)
    fmt.Printf("已使用: %d bytes (%.2f GB)\n", 
        quotaInfo.QuotaUsed, float64(quotaInfo.QuotaUsed)/1024/1024/1024)
    fmt.Printf("使用率: %.2f%%\n", 
        float64(quotaInfo.QuotaUsed)/float64(quotaInfo.Quota)*100)
    
    // 清理配额
    err = quota.CleanQuota("/data/user1")
    if err != nil {
        log.Fatal("清理配额失败:", err)
    }
    fmt.Println("配额清理成功")
}
```

### 2. 批量配额管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

type QuotaConfig struct {
    Path   string
    Size   string
    Inodes string
}

func main() {
    quota := api.NewQuotaManager()
    
    // 批量配置
    configs := []QuotaConfig{
        {Path: "/data/user1", Size: "10GiB", Inodes: "1000000"},
        {Path: "/data/user2", Size: "20GiB", Inodes: "2000000"},
        {Path: "/data/user3", Size: "5GiB", Inodes: "500000"},
    }
    
    // 批量设置配额
    for _, config := range configs {
        err := quota.SetQuota(config.Path, config.Size, config.Inodes)
        if err != nil {
            log.Printf("设置配额失败 %s: %v", config.Path, err)
            continue
        }
        fmt.Printf("成功设置配额: %s\n", config.Path)
    }
    
    // 批量查询配额
    for _, config := range configs {
        quotaInfo, err := quota.GetQuota(config.Path)
        if err != nil {
            log.Printf("查询配额失败 %s: %v", config.Path, err)
            continue
        }
        
        usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
        fmt.Printf("%s: 使用率 %.2f%%\n", config.Path, usagePercent)
    }
}
```

## 容器集成

### 1. Docker 容器配额管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

type DockerQuotaManager struct {
    quota *api.QuotaManager
}

func NewDockerQuotaManager() *DockerQuotaManager {
    return &DockerQuotaManager{
        quota: api.NewQuotaManager(),
    }
}

// SetContainerQuota 为容器设置配额
func (d *DockerQuotaManager) SetContainerQuota(containerID, size, inodes string) error {
    path := fmt.Sprintf("/var/lib/docker/containers/%s/data", containerID)
    return d.quota.SetQuota(path, size, inodes)
}

// GetContainerQuota 获取容器配额信息
func (d *DockerQuotaManager) GetContainerQuota(containerID string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/var/lib/docker/containers/%s/data", containerID)
    return d.quota.GetQuota(path)
}

// CleanContainerQuota 清理容器配额
func (d *DockerQuotaManager) CleanContainerQuota(containerID string) error {
    path := fmt.Sprintf("/var/lib/docker/containers/%s/data", containerID)
    return d.quota.CleanQuota(path)
}

// MonitorContainerUsage 监控容器使用情况
func (d *DockerQuotaManager) MonitorContainerUsage(containerID string) error {
    quotaInfo, err := d.GetContainerQuota(containerID)
    if err != nil {
        return fmt.Errorf("获取容器配额失败: %v", err)
    }
    
    usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
    
    fmt.Printf("容器 %s 配额使用情况:\n", containerID)
    fmt.Printf("  总配额: %.2f GB\n", float64(quotaInfo.Quota)/1024/1024/1024)
    fmt.Printf("  已使用: %.2f GB\n", float64(quotaInfo.QuotaUsed)/1024/1024/1024)
    fmt.Printf("  使用率: %.2f%%\n", usagePercent)
    
    if usagePercent > 80 {
        fmt.Printf("警告: 容器 %s 配额使用率超过 80%%\n", containerID)
    }
    
    return nil
}

func main() {
    manager := NewDockerQuotaManager()
    
    // 为容器设置配额
    err := manager.SetContainerQuota("abc123", "5GiB", "500000")
    if err != nil {
        log.Fatal("设置容器配额失败:", err)
    }
    
    // 监控容器使用情况
    err = manager.MonitorContainerUsage("abc123")
    if err != nil {
        log.Fatal("监控容器失败:", err)
    }
}
```

### 2. Kubernetes Pod 配额管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

type K8sQuotaManager struct {
    quota *api.QuotaManager
}

func NewK8sQuotaManager() *K8sQuotaManager {
    return &K8sQuotaManager{
        quota: api.NewQuotaManager(),
    }
}

// SetNamespaceQuota 为命名空间设置配额
func (k *K8sQuotaManager) SetNamespaceQuota(namespace, size, inodes string) error {
    path := fmt.Sprintf("/data/namespaces/%s", namespace)
    return k.quota.SetQuota(path, size, inodes)
}

// SetPodQuota 为 Pod 设置配额
func (k *K8sQuotaManager) SetPodQuota(namespace, podName, size, inodes string) error {
    path := fmt.Sprintf("/data/namespaces/%s/pods/%s", namespace, podName)
    return k.quota.SetQuota(path, size, inodes)
}

// GetNamespaceQuota 获取命名空间配额
func (k *K8sQuotaManager) GetNamespaceQuota(namespace string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/data/namespaces/%s", namespace)
    return k.quota.GetQuota(path)
}

// GetPodQuota 获取 Pod 配额
func (k *K8sQuotaManager) GetPodQuota(namespace, podName string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/data/namespaces/%s/pods/%s", namespace, podName)
    return k.quota.GetQuota(path)
}

// CleanPodQuota 清理 Pod 配额
func (k *K8sQuotaManager) CleanPodQuota(namespace, podName string) error {
    path := fmt.Sprintf("/data/namespaces/%s/pods/%s", namespace, podName)
    return k.quota.CleanQuota(path)
}

func main() {
    manager := NewK8sQuotaManager()
    
    // 为命名空间设置配额
    err := manager.SetNamespaceQuota("default", "100GiB", "10000000")
    if err != nil {
        log.Fatal("设置命名空间配额失败:", err)
    }
    
    // 为 Pod 设置配额
    err = manager.SetPodQuota("default", "nginx-pod", "10GiB", "1000000")
    if err != nil {
        log.Fatal("设置 Pod 配额失败:", err)
    }
    
    // 监控命名空间配额
    namespaceQuota, err := manager.GetNamespaceQuota("default")
    if err != nil {
        log.Fatal("获取命名空间配额失败:", err)
    }
    
    fmt.Printf("命名空间 default 配额使用情况:\n")
    fmt.Printf("  总配额: %.2f GB\n", float64(namespaceQuota.Quota)/1024/1024/1024)
    fmt.Printf("  已使用: %.2f GB\n", float64(namespaceQuota.QuotaUsed)/1024/1024/1024)
    
    // 监控 Pod 配额
    podQuota, err := manager.GetPodQuota("default", "nginx-pod")
    if err != nil {
        log.Fatal("获取 Pod 配额失败:", err)
    }
    
    fmt.Printf("Pod nginx-pod 配额使用情况:\n")
    fmt.Printf("  总配额: %.2f GB\n", float64(podQuota.Quota)/1024/1024/1024)
    fmt.Printf("  已使用: %.2f GB\n", float64(podQuota.QuotaUsed)/1024/1024/1024)
}
```

## 多租户平台集成

### 1. 云平台租户管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

type TenantQuotaManager struct {
    quota *api.QuotaManager
}

func NewTenantQuotaManager() *TenantQuotaManager {
    return &TenantQuotaManager{
        quota: api.NewQuotaManager(),
    }
}

// SetTenantQuota 为租户设置配额
func (t *TenantQuotaManager) SetTenantQuota(tenantID, size, inodes string) error {
    path := fmt.Sprintf("/data/tenants/%s", tenantID)
    return t.quota.SetQuota(path, size, inodes)
}

// GetTenantQuota 获取租户配额
func (t *TenantQuotaManager) GetTenantQuota(tenantID string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/data/tenants/%s", tenantID)
    return t.quota.GetQuota(path)
}

// MonitorTenantUsage 监控租户使用情况
func (t *TenantQuotaManager) MonitorTenantUsage(tenantID string) error {
    quotaInfo, err := t.GetTenantQuota(tenantID)
    if err != nil {
        return fmt.Errorf("获取租户配额失败: %v", err)
    }
    
    usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
    
    fmt.Printf("租户 %s 配额使用情况:\n", tenantID)
    fmt.Printf("  总配额: %.2f GB\n", float64(quotaInfo.Quota)/1024/1024/1024)
    fmt.Printf("  已使用: %.2f GB\n", float64(quotaInfo.QuotaUsed)/1024/1024/1024)
    fmt.Printf("  使用率: %.2f%%\n", usagePercent)
    
    // 根据使用率发送告警
    if usagePercent > 90 {
        fmt.Printf("严重告警: 租户 %s 配额使用率超过 90%%\n", tenantID)
    } else if usagePercent > 80 {
        fmt.Printf("警告: 租户 %s 配额使用率超过 80%%\n", tenantID)
    }
    
    return nil
}

func main() {
    manager := NewTenantQuotaManager()
    
    // 为不同租户设置配额
    tenants := map[string]string{
        "tenant-a": "100GiB",
        "tenant-b": "50GiB",
        "tenant-c": "200GiB",
    }
    
    for tenantID, size := range tenants {
        err := manager.SetTenantQuota(tenantID, size, "10000000")
        if err != nil {
            log.Printf("设置租户 %s 配额失败: %v", tenantID, err)
            continue
        }
        fmt.Printf("成功设置租户 %s 配额\n", tenantID)
    }
    
    // 监控所有租户
    for tenantID := range tenants {
        err := manager.MonitorTenantUsage(tenantID)
        if err != nil {
            log.Printf("监控租户 %s 失败: %v", tenantID, err)
        }
    }
}
```

## CI/CD 构建系统集成

### 1. Jenkins 构建配额管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

type JenkinsQuotaManager struct {
    quota *api.QuotaManager
}

func NewJenkinsQuotaManager() *JenkinsQuotaManager {
    return &JenkinsQuotaManager{
        quota: api.NewQuotaManager(),
    }
}

// SetBuildQuota 为构建任务设置配额
func (j *JenkinsQuotaManager) SetBuildQuota(buildID, size, inodes string) error {
    path := fmt.Sprintf("/var/lib/jenkins/workspace/build-%s", buildID)
    return j.quota.SetQuota(path, size, inodes)
}

// GetBuildQuota 获取构建任务配额
func (j *JenkinsQuotaManager) GetBuildQuota(buildID string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/var/lib/jenkins/workspace/build-%s", buildID)
    return j.quota.GetQuota(path)
}

// CleanBuildQuota 清理构建任务配额
func (j *JenkinsQuotaManager) CleanBuildQuota(buildID string) error {
    path := fmt.Sprintf("/var/lib/jenkins/workspace/build-%s", buildID)
    return j.quota.CleanQuota(path)
}

// MonitorBuildUsage 监控构建使用情况
func (j *JenkinsQuotaManager) MonitorBuildUsage(buildID string) error {
    quotaInfo, err := j.GetBuildQuota(buildID)
    if err != nil {
        return fmt.Errorf("获取构建配额失败: %v", err)
    }
    
    usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
    
    fmt.Printf("构建任务 %s 配额使用情况:\n", buildID)
    fmt.Printf("  总配额: %.2f GB\n", float64(quotaInfo.Quota)/1024/1024/1024)
    fmt.Printf("  已使用: %.2f GB\n", float64(quotaInfo.QuotaUsed)/1024/1024/1024)
    fmt.Printf("  使用率: %.2f%%\n", usagePercent)
    
    if usagePercent > 90 {
        fmt.Printf("警告: 构建任务 %s 配额使用率超过 90%%，可能影响构建\n", buildID)
    }
    
    return nil
}

func main() {
    manager := NewJenkinsQuotaManager()
    
    // 为构建任务设置配额
    err := manager.SetBuildQuota("12345", "5GiB", "500000")
    if err != nil {
        log.Fatal("设置构建配额失败:", err)
    }
    
    // 监控构建使用情况
    err = manager.MonitorBuildUsage("12345")
    if err != nil {
        log.Fatal("监控构建失败:", err)
    }
    
    // 构建完成后清理配额
    err = manager.CleanBuildQuota("12345")
    if err != nil {
        log.Fatal("清理构建配额失败:", err)
    }
    fmt.Println("构建配额清理成功")
}
```

## 机器学习训练数据管理

### 1. 训练任务缓存管理

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

type MLQuotaManager struct {
    quota *api.QuotaManager
}

func NewMLQuotaManager() *MLQuotaManager {
    return &MLQuotaManager{
        quota: api.NewQuotaManager(),
    }
}

// SetTrainingQuota 为训练任务设置配额
func (m *MLQuotaManager) SetTrainingQuota(jobID, size, inodes string) error {
    path := fmt.Sprintf("/cache/job-%s", jobID)
    return m.quota.SetQuota(path, size, inodes)
}

// GetTrainingQuota 获取训练任务配额
func (m *MLQuotaManager) GetTrainingQuota(jobID string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/cache/job-%s", jobID)
    return m.quota.GetQuota(path)
}

// CleanTrainingQuota 清理训练任务配额
func (m *MLQuotaManager) CleanTrainingQuota(jobID string) error {
    path := fmt.Sprintf("/cache/job-%s", jobID)
    return m.quota.CleanQuota(path)
}

// MonitorTrainingUsage 监控训练使用情况
func (m *MLQuotaManager) MonitorTrainingUsage(jobID string) error {
    quotaInfo, err := m.GetTrainingQuota(jobID)
    if err != nil {
        return fmt.Errorf("获取训练配额失败: %v", err)
    }
    
    usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
    
    fmt.Printf("训练任务 %s 缓存使用情况:\n", jobID)
    fmt.Printf("  总配额: %.2f GB\n", float64(quotaInfo.Quota)/1024/1024/1024)
    fmt.Printf("  已使用: %.2f GB\n", float64(quotaInfo.QuotaUsed)/1024/1024/1024)
    fmt.Printf("  使用率: %.2f%%\n", usagePercent)
    
    if usagePercent > 95 {
        fmt.Printf("严重警告: 训练任务 %s 缓存使用率超过 95%%，可能影响训练\n", jobID)
    }
    
    return nil
}

func main() {
    manager := NewMLQuotaManager()
    
    // 为训练任务设置缓存配额
    err := manager.SetTrainingQuota("12345", "20GiB", "1000000")
    if err != nil {
        log.Fatal("设置训练配额失败:", err)
    }
    
    // 定期监控训练使用情况
    err = manager.MonitorTrainingUsage("12345")
    if err != nil {
        log.Fatal("监控训练失败:", err)
    }
    
    // 训练完成后清理缓存配额
    err = manager.CleanTrainingQuota("12345")
    if err != nil {
        log.Fatal("清理训练配额失败:", err)
    }
    fmt.Println("训练缓存配额清理成功")
}
```

## 错误处理最佳实践

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

func main() {
    quota := api.NewQuotaManager()
    
    // 设置配额时进行错误处理
    err := quota.SetQuota("/data/user1", "10GiB", "1000000")
    if err != nil {
        // 根据错误类型进行不同处理
        switch {
        case err.Error() == "not supported":
            log.Fatal("文件系统不支持项目配额，请检查 XFS 配置")
        case err.Error() == "permission denied":
            log.Fatal("权限不足，请使用 sudo 运行")
        default:
            log.Fatal("设置配额失败:", err)
        }
    }
    
    // 查询配额时进行错误处理
    quotaInfo, err := quota.GetQuota("/data/user1")
    if err != nil {
        if err.Error() == "path not found":
            log.Fatal("路径不存在，请检查路径是否正确")
        }
        log.Fatal("查询配额失败:", err)
    }
    
    // 检查配额使用情况
    if quotaInfo.Quota > 0 {
        usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
        if usagePercent > 90 {
            log.Printf("警告: 配额使用率超过 90%%")
        }
    }
}
```

这些示例展示了如何在不同场景下使用 XFS Quota Manager API，包括容器管理、Kubernetes 集成、多租户平台、CI/CD 系统和机器学习训练等场景。 