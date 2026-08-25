# File 模块主链路

本文档记录 `internal/file` 的主要业务链路。File 模块只负责业务上传入口和对象 URL 生成，真正的 MinIO client 由 `internal/pkg/miniox` 初始化后注入。

## 模块启动链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant Module as file.Module
  participant Logic as file.logic
  participant Handler as file.handler
  participant Storage as MinIO client

  Router->>Module: NewModule(Config, Storage, Log)
  Module->>Logic: NewFileLogic(Storage, Config, Log)
  Module->>Handler: NewFileHandler(logic, Log)
  Router->>Module: RegisterRoutes(api.Group('/file'))
  Logic-->>Storage: use injected object storage interface
```

## 文件上传链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as FileHandler
  participant Logic as FileLogic
  participant Storage as MinIO

  Client->>Handler: POST /api/file/upload
  Handler->>Handler: auth user and read multipart file
  Handler->>Logic: Upload(ctx, file, header)
  Logic->>Logic: check minio enable, header, file size
  Logic->>Logic: stream file to md5 hash
  Logic->>Logic: seek file back to start
  Logic->>Logic: build object name hash + original extension
  Logic->>Storage: StatObject(bucket, objectName)
  alt object already exists
    Storage-->>Logic: object info
    Logic-->>Handler: existing object URL
  else object missing
    Logic->>Storage: PutObject(bucket, objectName, file)
    Storage-->>Logic: upload info
    Logic-->>Handler: new object URL
  end
  Handler-->>Client: response
```

## 边界说明

- File 模块不处理具体对象存储初始化，不读取 `global`。
- 上传去重依据文件内容 MD5，同内容文件复用同一个对象名。
- 文件大小、默认路径等固定策略放在 `internal/file/constant`。
