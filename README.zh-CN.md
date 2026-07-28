# Hometab

> 以单一可执行文件交付的本地优先个人起始页。

[![CI](https://github.com/ccviolett/Hometab/actions/workflows/ci.yml/badge.svg)](https://github.com/ccviolett/Hometab/actions/workflows/ci.yml)
[![Pages](https://github.com/ccviolett/Hometab/actions/workflows/pages.yml/badge.svg)](https://ccviolett.github.io/Hometab/)

[项目主页](https://ccviolett.github.io/Hometab/?lang=zh-CN) · [English](./README.md)

![Hometab 分组链接与搜索界面](./site/assets/hometab-demo.png)

Hometab 将链接分组、多引擎搜索、固定 HTTP 操作、壁纸、图标管理和版本化备份集中在一个自托管应用中。React 前端内嵌于 Go 服务，所有持久化数据保存在本地 SQLite 数据库中。

## 功能

- **链接与流程**：分组、拖拽排序，以及可复用的链接流程。
- **搜索**：Google、Bing、DuckDuckGo 和自定义搜索引擎。
- **固定请求**：结构化请求头、查询参数、Body 和 JSON Path 结果提取。
- **请求安全**：URL、DNS、重定向、超时、响应大小和并发限制。
- **壁纸**：Bing 每日/随机、自定义 URL、本地回退和缓存。
- **域名图标**：自动发现、上传、替换、重试、冲突处理和本地存储。
- **备份恢复**：版本化 ZIP 导出、合并导入、覆盖恢复和自动恢复点。
- **双语界面**：简体中文和 English。
- **专注模式**：双击壁纸，在仪表盘与仅搜索视图之间切换。

## 构建要求

从源码构建需要：

- `apps/backend-go-fiber/go.mod` 中声明的 Go 版本
- Node.js 22 或更高版本
- npm

## 构建并运行

```bash
git clone https://github.com/ccviolett/Hometab.git
cd Hometab
make setup
make build
./apps/backend-go-fiber/bin/hometab
```

服务就绪后，Hometab 会自动使用系统默认浏览器打开应用。默认地址是：

```text
http://127.0.0.1:52173
```

默认端口被占用时，服务会自动尝试后续端口，并打开实际使用的地址。在后台或自动化环境中可使用 `--no-open` 禁止打开浏览器。

## 数据位置

可通过 `HOME_DATABASE_PATH` 覆盖数据库路径。默认位置：

- macOS：`~/Library/Application Support/Hometab/data.db`
- Linux：`~/.config/Hometab/data.db`
- Windows：`%AppData%\Hometab\data.db`

已有的旧 `Home` 目录会被自动识别，无需手工迁移。

## macOS 后台服务

```bash
./apps/backend-go-fiber/bin/hometab --install
./apps/backend-go-fiber/bin/hometab --status
./apps/backend-go-fiber/bin/hometab --stop
./apps/backend-go-fiber/bin/hometab --start
./apps/backend-go-fiber/bin/hometab --uninstall
```

卸载服务不会删除用户数据。

## 开发模式

```bash
# 终端 1
cd apps/backend-go-fiber
make run-dev

# 终端 2
cd apps/frontend-react-vite
npm run dev
```

提交前运行完整质量门槛：

```bash
make check
```

## 安全边界

Hometab 是无内置认证的可信单用户应用。请保持默认的本机监听地址。远程访问必须使用启用 HTTPS 和身份认证的反向代理。

当服务监听非 loopback 地址时，固定请求执行默认关闭。完整说明见 [SECURITY.md](./SECURITY.md)。

## 文档

- [架构](./docs/ARCHITECTURE.md)
- [开发指南](./docs/DEVELOPMENT.md)
- [贡献指南](./CONTRIBUTING.md)
- [安全策略](./SECURITY.md)
- [变更记录](./CHANGELOG.md)

## 发布

创建并推送语义化发布标签：

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions 会构建 macOS、Linux 和 Windows 压缩包，生成 SHA-256 校验和并发布 GitHub Release。

## License

[MIT](./LICENSE)
