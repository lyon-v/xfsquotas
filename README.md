# XFS Quota Manager

一个用于管理 XFS 文件系统项目配额的 Go 工具库，支持通过命令行和编程接口进行配额管理。

## 功能特性

- 🚀 **高性能**: 基于系统调用直接操作 XFS 配额
- 🔧 **易集成**: 提供简洁的 Go API 接口
- 📦 **容器友好**: 支持容器环境下的配额管理
- 🛡️ **安全可靠**: 支持配额文件持久化存储
- 📊 **实时监控**: 支持配额使用情况查询

## 安装

```bash
# 克隆项目
git clone https://github.com/your-repo/xfsquotas.git
cd xfsquotas

# 构建
go build -o xfsquota ./cmd/xfsquota

# 安装到系统
sudo cp xfsquota /usr/local/bin/
```

## 使用场景

### 1. 容器 rootfs 或数据目录限额

**目标**: 防止单个容器占用过多磁盘空间

**实现**: 在容器创建时自动为其挂载目录设置项目 ID 和限额

**集成方式**: 可嵌入 kubelet 容器创建流程中（通过 CRI 插件或 kubelet hook）

```bash
# 为容器数据目录设置 10GB 限额
xfsquota set /var/lib/docker/containers/abc123/data -s 10GiB -i 1000000

# 查询容器配额使用情况
xfsquota get /var/lib/docker/containers/abc123/data
```

### 2. 多租户平台的存储隔离

**目标**: 为不同用户/租户/namespace 分配不同目录的存储限额，防止"爆盘"

**场景**:
- 云平台租户隔离（如 AI 云平台）
- DevOps 平台的项目目录管理

**优势**: 项目配额粒度比普通用户/组配额更细，可作用于单个目录

```bash
# 为租户 A 设置 100GB 存储限额
xfsquota set /data/tenant-a -s 100GiB -i 5000000

# 为租户 B 设置 50GB 存储限额
xfsquota set /data/tenant-b -s 50GiB -i 2500000

# 监控租户存储使用情况
xfsquota get /data/tenant-a
xfsquota get /data/tenant-b
```

### 3. 大规模机器学习训练数据缓存管理

**目标**: 为每个训练任务的中间数据目录设定存储限制

**用途**:
- 避免缓存膨胀导致系统盘满
- 可定期通过 clean 命令清理任务已结束目录

```bash
# 为训练任务设置 20GB 缓存限额
xfsquota set /cache/job-12345 -s 20GiB -i 1000000

# 任务结束后清理配额
xfsquota clean /cache/job-12345
```

### 4. 构建系统的磁盘使用限制

**目标**: 在 CI/CD 构建集群中，限制每个构建任务的磁盘使用上限

**示例**:
- GitLab Runner 本地构建缓存目录限制
- Jenkins agent 中的 workspace 限额控制

```bash
# 为构建任务设置 5GB 工作空间限额
xfsquota set /var/lib/jenkins/workspace/build-123 -s 5GiB -i 500000

# 构建完成后清理
xfsquota clean /var/lib/jenkins/workspace/build-123
```

## 命令行使用

### 基本命令

```bash
# 查看帮助
xfsquota --help

# 查询配额信息
xfsquota get <path>

# 设置配额
xfsquota set <path> -s <size> -i <inodes>

# 清理配额
xfsquota clean <path>
```

### 使用示例

```bash
# 为 /data/user1 设置 10GB 和 100万 inode 限额
xfsquota set /data/user1 -s 10GiB -i 1000000

# 查询配额使用情况
xfsquota get /data/user1
# 输出示例:
# quota Size(bytes): 10737418240
# quota Inodes: 1000000
# diskUsage Size(bytes): 2147483648
# diskUsage Inodes: 150000

# 清理配额
xfsquota clean /data/user1
```

## 编程接口使用

### 基本用法

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
    
    // 查询配额
    quotaInfo, err := quota.GetQuota("/data/user1")
    if err != nil {
        log.Fatal("查询配额失败:", err)
    }
    
    fmt.Printf("配额大小: %d bytes\n", quotaInfo.Quota)
    fmt.Printf("已使用: %d bytes\n", quotaInfo.QuotaUsed)
    fmt.Printf("Inode 限额: %d\n", quotaInfo.Inodes)
    fmt.Printf("已使用 Inode: %d\n", quotaInfo.InodesUsed)
    
    // 清理配额
    err = quota.CleanQuota("/data/user1")
    if err != nil {
        log.Fatal("清理配额失败:", err)
    }
}
```

### 容器集成示例

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

// ContainerQuotaManager 容器配额管理器
type ContainerQuotaManager struct {
    quota *api.QuotaManager
}

func NewContainerQuotaManager() *ContainerQuotaManager {
    return &ContainerQuotaManager{
        quota: api.NewQuotaManager(),
    }
}

// SetContainerQuota 为容器设置配额
func (c *ContainerQuotaManager) SetContainerQuota(containerID, size, inodes string) error {
    path := fmt.Sprintf("/var/lib/docker/containers/%s/data", containerID)
    return c.quota.SetQuota(path, size, inodes)
}

// GetContainerQuota 获取容器配额信息
func (c *ContainerQuotaManager) GetContainerQuota(containerID string) (*api.DiskQuotaSize, error) {
    path := fmt.Sprintf("/var/lib/docker/containers/%s/data", containerID)
    return c.quota.GetQuota(path)
}

// CleanContainerQuota 清理容器配额
func (c *ContainerQuotaManager) CleanContainerQuota(containerID string) error {
    path := fmt.Sprintf("/var/lib/docker/containers/%s/data", containerID)
    return c.quota.CleanQuota(path)
}

func main() {
    manager := NewContainerQuotaManager()
    
    // 为容器设置 5GB 配额
    err := manager.SetContainerQuota("abc123", "5GiB", "500000")
    if err != nil {
        log.Fatal("设置容器配额失败:", err)
    }
    
    // 监控容器配额使用情况
    quotaInfo, err := manager.GetContainerQuota("abc123")
    if err != nil {
        log.Fatal("获取容器配额失败:", err)
    }
    
    usagePercent := float64(quotaInfo.QuotaUsed) / float64(quotaInfo.Quota) * 100
    fmt.Printf("容器配额使用率: %.2f%%\n", usagePercent)
}
```

### Kubernetes 集成示例

```go
package main

import (
    "fmt"
    "log"
    "xfsquotas/api"
)

// K8sQuotaManager Kubernetes 配额管理器
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
}
```

## 系统要求

- Linux 系统
- XFS 文件系统
- 内核支持项目配额
- 需要 root 权限或适当的权限

## 配置

### 启用 XFS 项目配额

```bash
# 检查文件系统是否支持项目配额
mount | grep xfs

# 重新挂载文件系统以启用项目配额
mount -o remount,prjquota /dev/sda1 /data

# 或者修改 /etc/fstab
echo "/dev/sda1 /data xfs defaults,prjquota 0 0" >> /etc/fstab
```

## 故障排除

### 常见问题

1. **权限不足**
   ```bash
   # 错误: permission denied
   # 解决: 使用 sudo 运行命令
   sudo xfsquota set /data/user1 -s 10GiB -i 1000000
   ```

2. **文件系统不支持**
   ```bash
   # 错误: not supported
   # 解决: 确保使用 XFS 文件系统并启用项目配额
   mount -o remount,prjquota /dev/sda1 /data
   ```

3. **路径不存在**
   ```bash
   # 错误: path is required
   # 解决: 确保指定正确的路径
   xfsquota get /data/user1
   ```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

Apache License 2.0 