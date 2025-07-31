# XFS Quota 项目重构

## 重构目标

1. 使用 `urfave/cli` 替代 `cobra` 作为 CLI 框架
2. 使用项目自己的库替代外部依赖 `github.com/silenceper/xfsquota/pkg/projectquota`
3. 采用更清晰的目录结构

## 新的目录结构

```
xfsquotas/
├── cmd/                                # 各命令入口（多命令结构）
│   └── xfsquota/
│       └── main.go                     # 主程序入口，注册 CLI 命令
├── internal/                           # 内部逻辑，非导出 API
│   ├── cli/                            # CLI 命令定义和注册
│   │   ├── get.go                      # get 子命令逻辑
│   │   ├── set.go                      # set 子命令逻辑
│   │   └── clean.go                    # clean 子命令逻辑
│   ├── mount/                          # 与挂载信息相关的逻辑
│   │   └── parser.go                   # 原 mountpoint.go，挂载点解析工具
│   ├── project/                        # XFS 项目配额核心逻辑
│   │   ├── quota.go                    # 核心实现，set/query/clean 等
│   │   ├── file.go                     # 管理 /etc/projects 和 /etc/projid
│   │   └── types.go                    # 定义 DiskQuotaSize 等结构体
│   └── xfs/                            # 对 xfs_quota 命令的封装
│       └── client.go                   # 封装系统调用/命令执行
├── api/                                # 可导出的 Go SDK（可选，若计划做库）
│   └── quota.go                        # 暴露对外使用的 API 接口
├── dos/                                # 文档目录
├── scripts/                            # 启动脚本或辅助脚本
├── test/                               # 单元测试或集成测试目录（可选）
├── Makefile                            # 构建脚本
├── go.mod
├── go.sum
├── .gitignore
├── LICENSE
└── xfsquota                            # 编译输出
```

## 主要变更

### 1. CLI 框架变更
- **从**: `github.com/spf13/cobra`
- **到**: `github.com/urfave/cli/v2`
- **原因**: urfave/cli 更轻量，API 更简洁

### 2. 依赖管理
- **移除**: `github.com/silenceper/xfsquota/pkg/projectquota`
- **使用**: 项目内部的 `internal/project` 包
- **原因**: 减少外部依赖，提高可控性

### 3. 目录结构优化
- **internal/**: 存放内部实现，不对外暴露
- **api/**: 提供对外使用的 SDK 接口
- **cmd/**: 保持命令行入口的简洁性
- **scripts/**: 构建和部署脚本
- **test/**: 测试代码（预留）

## 文件功能说明

### 核心文件
- `internal/project/quota.go`: XFS 项目配额的核心实现
- `internal/mount/parser.go`: 挂载点解析工具
- `internal/project/file.go`: 配额文件管理
- `internal/project/types.go`: 数据结构定义

### CLI 命令
- `internal/cli/get.go`: 查询配额信息
- `internal/cli/set.go`: 设置配额信息
- `internal/cli/clean.go`: 清理配额信息

### API 层
- `api/quota.go`: 对外暴露的高级接口

## 构建和运行

```bash
# 构建
go build -o xfsquota ./cmd/xfsquota

# 运行
./xfsquota --help
./xfsquota get /path/to/directory
./xfsquota set /path/to/directory -s 100MiB -i 1000
./xfsquota clean /path/to/directory
```

## 优势

1. **更清晰的架构**: 内部实现与外部 API 分离
2. **更好的可维护性**: 模块化设计，职责明确
3. **更少的依赖**: 移除外部依赖，提高稳定性
4. **更好的扩展性**: 预留了测试和脚本目录
5. **更现代的 CLI**: 使用 urfave/cli 提供更好的用户体验
