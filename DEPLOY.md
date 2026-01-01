# Nurture 项目部署指南

## � 极简部署 (推荐)

最简单的方式是使用 Docker Compose 一键启动所有服务（应用 + 数据库 + Redis）。

### 1. 前置要求
- 安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/) 或 Docker Engine + Docker Compose。

### 2. 一键启动
在项目根目录下执行：

```bash
docker-compose up -d --build
```

### 3. 验证
服务启动后，API 将在 `http://localhost:8080` 上可用。

---

## ⚙️ 自定义配置 (可选)

默认情况下，Docker 镜像会使用 `internal/etc/template.yaml` 作为默认配置。如果您需要修改生产环境配置（如数据库密码、密钥等）：

1.  在本地修改 `internal/etc/template.yaml`（或者创建一个 `local.yaml`，Docker 构建时会优先使用它，但注意不要提交敏感信息到 git）。
2.  或者，您可以挂载配置文件到容器中：

修改 `docker-compose.yaml`:
```yaml
  nurture-api:
    # ...
    volumes:
      - ./internal/etc/local.yaml:/app/internal/etc/local.yaml
```

---

## � 其他部署方式

### 手动构建与部署

如果您不使用 Docker，可以参考以下步骤手动部署。

#### 1. 编译
```bash
go build -o nurture internal/main.go
```

#### 2. 配置
确保运行目录下有配置文件：
```bash
cp internal/etc/template.yaml internal/etc/local.yaml
```

#### 3. 运行
```bash
./nurture
```
