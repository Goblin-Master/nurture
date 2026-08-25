# AI 模块主链路

本文档记录 `internal/ai` 的主要业务链路。AI 模块负责 HTTP 入口、对话编排、知识库上传、成长分析和成长报告；模型、Embedding、向量库与历史存储能力由业务无关的 `internal/pkg/aix` 提供。

## 模块启动链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant Module as ai.Module
  participant Repo as ai.repo
  participant Logic as ai.logic
  participant Handler as ai.handler
  participant AIX as pkg/aix
  participant Baby as baby.Client adapter

  Router->>Module: NewModule(RDB, Log, AIX, AIConfig, DBEnabled, GrowthReader)
  Module->>Repo: NewAIRepo(AIX, RDB, Log)
  Module->>Logic: NewAILogic(repo, AIX, AIConfig, GrowthReader, DBEnabled, Log)
  Module->>Handler: NewAIHandler(logic, Log)
  Router->>Module: RegisterRoutes(api.Group('/ai'))
  Module-->>Router: register knowledge, chat, growth routes
  Logic-->>AIX: use chat, embedding, vector, history capabilities
  Logic-->>Baby: read baby profile and growth records through injected boundary
```

## 对话流链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as AIHandler
  participant Logic as AILogic
  participant Repo as AIRepo
  participant Redis as Redis
  participant AIX as pkg/aix
  participant Vector as pgvector
  participant Baby as baby.Client adapter
  participant LLM as chat provider

  Client->>Handler: POST /api/ai/chat/stream
  Handler->>Handler: auth user and bind ChatStreamReq
  Handler->>Logic: ChatStream(ctx, userID, req, stream)
  Logic->>Repo: GetRecentHistory(userID, sessionID, context limit)
  Repo->>Redis: load recent chat history
  Redis-->>Repo: history messages
  Repo-->>Logic: history

  alt auto_context enabled with baby_id
    Logic->>Baby: GetBabyByIDAndUser(userID, babyID)
    Logic->>Baby: ListGrowthRecordsByBabyIDBetween(babyID, from, to)
    Baby-->>Logic: baby profile and records
    Logic->>Logic: build extra_context JSON and local trend analysis
  end

  alt KB enabled and embedding available
    Logic->>Repo: SimilaritySearch(message, collections, topK)
    Repo->>AIX: SimilaritySearch
    AIX->>Vector: embed query and search collections
    Vector-->>AIX: matched documents
    AIX-->>Repo: documents
    Repo-->>Logic: documents
    Logic->>Logic: build rag_context
  end

  Logic->>AIX: BuildMessages(history, user message, images, rag_context, extra_context)
  Logic->>AIX: StreamChat(messages)
  AIX->>LLM: streaming GenerateContent
  loop chunks
    LLM-->>AIX: content chunk
    AIX-->>Logic: chunk
    Logic-->>Handler: SSE content event
    Handler-->>Client: event: message
  end
  AIX-->>Logic: full assistant response
  Logic->>Repo: SaveMessage(user)
  Logic->>Repo: SaveMessage(assistant)
  Repo->>Redis: append full history with TTL
  Logic-->>Handler: SSE done event
  Handler-->>Client: done
```

## 知识库上传链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as AIHandler
  participant Logic as AILogic
  participant Repo as AIRepo
  participant AIX as pkg/aix
  participant Vector as pgvector

  Client->>Handler: POST /api/ai/knowledge/upload
  Handler->>Handler: auth user and bind KnowledgeUploadReq
  Handler->>Logic: UploadKnowledge(ctx, userID, req)
  Logic->>Logic: map space_type to collection name
  Logic->>Logic: require embedding and DB availability
  Logic->>Repo: AddDocument(collection, content)
  Repo->>AIX: AddDocument
  AIX->>AIX: split content into chunks
  AIX->>AIX: embed chunks with selected embedding provider
  AIX->>Vector: store documents in collection
  Vector-->>AIX: ok
  AIX-->>Repo: nil
  Repo-->>Logic: nil
  Logic-->>Handler: nil
  Handler-->>Client: response
```

## 成长报告链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as AIHandler
  participant Logic as AILogic
  participant Baby as baby.Client adapter
  participant AIX as pkg/aix
  participant LLM as chat provider

  Client->>Handler: POST /api/ai/report/growth
  Handler->>Logic: GrowthReport(ctx, userID, req)
  Logic->>Baby: GetBabyByIDAndUser(userID, babyID)
  Logic->>Baby: ListGrowthRecordsByBabyIDBetween(babyID, from, to)
  Baby-->>Logic: profile and growth records
  Logic->>Logic: calculate height, weight, head circumference trends
  alt chat model available
    Logic->>AIX: Build growth report prompts
    Logic->>AIX: StreamChat(messages)
    AIX->>LLM: generate Markdown
    LLM-->>AIX: Markdown
    AIX-->>Logic: Markdown
  else chat disabled or generation failed
    Logic->>Logic: build fallback Markdown from local analysis
  end
  Logic-->>Handler: GrowthReportResp(data, markdown)
  Handler-->>Client: response
```

## 边界说明

- AI 模块不直接初始化模型或外部客户端，`AIX`、Redis、日志和 baby 读取能力都由上层注入；当前 baby 读取由 router 中的 adapter 基于 `baby.Client` 提供。
- `chat.enable` 与 `embedding.enable` 分开判断：普通对话、知识库上传、RAG 检索和成长报告可以按能力降级。
- 知识库固定策略放在 `internal/config` 的 AI 配置与 `internal/ai/constant` 中；DTO 不承载配置。
