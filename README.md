# data_analysis_server

数据分析服务后端，基于 Gin + GORM + Casbin。

## 技术栈

- Go 1.24
- Gin、GORM、Casbin、JWT、Swagger

## 功能

- 用户 / 角色 / 菜单 / 部门 / 岗位管理
- JWT 认证、Casbin RBAC 权限
- 操作日志、登录日志
- 定时任务
- 代码生成
- 业务库（与系统库分离，仅查询）

## 目录

```
app/          业务模块（admin / jobs / other）
cmd/          命令：server、migrate、config、version、app
common/       中间件、数据库、通用逻辑
config/       配置文件
docs/         Swagger 文档
```

## 环境

- Go 1.24+

## 配置

主配置：`config/settings.yml`

- `settings.database`：系统库（权限、用户等）
- `settings.extend.business`：业务库，仅查询，不参与权限/租户分库

## 启动

```bash
go mod tidy
go build -o go-admin .

# 首次需要初始化系统库
./go-admin migrate -c config/settings.yml

# 启动 API
./go-admin server -c config/settings.yml
```

Windows：

```bash
go-admin.exe migrate -c config/settings.yml
go-admin.exe server -c config/settings.yml
```

无参数运行时默认启动 API，便于 IDE 一键调试。启动时加 `-a true` 会自动补齐缺失的接口数据。

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `server` | 启动 API |
| `migrate` | 初始化 / 迁移系统库 |
| `config` | 打印当前配置 |
| `version` | 查看版本 |
| `app` | 生成新业务模块 |

## Docker

```bash
docker build -t go-admin .
docker run --name go-admin -p 8000:8000 -v ./config:/config -d go-admin
```

也可使用仓库根目录的 `docker-compose.yml`。

## License

[MIT](LICENSE.md)
