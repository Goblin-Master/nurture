# Chat 模块主链路

本文档记录 `internal/chat` 的主要业务链路。所有图使用 Mermaid `sequenceDiagram`，通过 `participant` 表达泳道。

## 模块启动链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant Module as chat.Module
  participant Repo as chat.repo
  participant Logic as chat.logic
  participant Handler as chat.handler
  participant Hub as chat.session.Hub
  participant Consumer as chat.worker.Worker
  participant Outbox as chat.worker.OutboxWorker
  participant MQ as RabbitMQ

  Router->>Module: NewModule(DB, RDB, RMQ, Log, middleware)
  Module->>Hub: NewHub()
  Module->>Hub: Run()
  Module->>Repo: NewChatRepo(DB, RDB, Log)
  Module->>Logic: NewChatLogic(repo, limiter)
  Module->>Handler: NewChatHandler(logic, hub)
  Module->>Consumer: NewWorker(RMQ, hub, Log)
  Consumer->>MQ: Consume(chat.event, chat.direct/chat.group)
  Module->>Outbox: NewOutboxWorker(repo, RMQ, Log)
  Outbox->>MQ: DeclareTopicExchange(chat.event)
  Router->>Module: RegisterRoutes(api.Group('/chat'), ws)
  Module-->>Router: register HTTP and WS routes
```

## 私聊发送链路

```mermaid
sequenceDiagram
  autonumber
  participant Sender as sender client
  participant Handler as ChatHandler
  participant Logic as ChatLogic
  participant Repo as ChatRepo
  participant DB as PostgreSQL
  participant Outbox as OutboxWorker
  participant MQ as RabbitMQ
  participant Consumer as Worker
  participant Hub as session.Hub
  participant Receiver as receiver client

  Sender->>Handler: WS /ws/chat?token=...&user_id=receiver
  Handler->>Handler: parse token and register direct client
  Sender->>Handler: send {"op":"send","message_id","type","content"}
  Handler->>Logic: HandleDirectMessage(ctx, senderID, receiverID, message)
  Logic->>Logic: parse JSON and validate message
  Logic->>Logic: rate limit sender and sender/receiver pair
  Logic->>Logic: build DirectMessage event payload
  Logic->>Repo: SaveDirectMessage(..., outbox)
  Repo->>DB: begin transaction
  Repo->>DB: insert chat_direct_message
  Repo->>DB: insert chat_event_outbox
  Repo->>DB: commit
  Repo-->>Logic: created, nil
  Logic-->>Handler: DirectAckMessage(ok=true)
  Handler-->>Sender: ack

  loop poll outbox
    Outbox->>Repo: ListPendingOutbox(now, staleBefore, limit)
    Repo->>DB: claim pending event as publishing
    DB-->>Repo: direct event
    Repo-->>Outbox: event
    Outbox->>MQ: publish chat.direct
    MQ-->>Outbox: publisher confirm
    Outbox->>Repo: MarkOutboxPublished
  end

  MQ->>Consumer: deliver chat.direct event
  Consumer->>Hub: DeliverToUser(direct, receiverID, eventID, payload)
  Hub-->>Receiver: new_message
```

## 群聊发送链路

```mermaid
sequenceDiagram
  autonumber
  participant Sender as sender client
  participant Handler as ChatHandler
  participant Logic as ChatLogic
  participant Repo as ChatRepo
  participant DB as PostgreSQL
  participant Outbox as OutboxWorker
  participant MQ as RabbitMQ
  participant Consumer as Worker
  participant Hub as session.Hub
  participant Members as subscribed room members

  Sender->>Handler: WS /ws/group?token=...
  Handler->>Handler: parse token and register group client
  Sender->>Handler: send {"op":"send","group_id","message_id","type","content"}
  Handler->>Logic: HandleGroupMessage(ctx, senderID, message)
  Logic->>Logic: route op=send to handleGroupSend
  Logic->>Logic: validate fields and message type
  Logic->>Repo: GetMemberRole(groupID, senderID)
  Repo->>DB: select chat_group_member.role
  DB-->>Repo: member role
  Repo-->>Logic: role
  Logic->>Logic: rate limit sender and sender/group pair
  Logic->>Logic: build GroupMessage event payload
  Logic->>Repo: SaveMessage(..., outbox)
  Repo->>DB: begin transaction
  Repo->>DB: insert chat_group_message
  Repo->>DB: insert chat_event_outbox
  Repo->>DB: commit
  Repo-->>Logic: created, nil
  Logic-->>Handler: GroupAckMessage(ok=true)
  Handler-->>Sender: ack

  loop poll outbox
    Outbox->>Repo: ListPendingOutbox(now, staleBefore, limit)
    Repo->>DB: claim pending event as publishing
    DB-->>Repo: group event
    Repo-->>Outbox: event
    Outbox->>MQ: publish chat.group
    MQ-->>Outbox: publisher confirm
    Outbox->>Repo: MarkOutboxPublished
  end

  MQ->>Consumer: deliver chat.group event
  Consumer->>Hub: DeliverToRoom(groupID, eventID, payload)
  Hub-->>Members: new_message
```

## Outbox 投递与重试链路

```mermaid
sequenceDiagram
  autonumber
  participant Outbox as OutboxWorker
  participant Repo as ChatRepo
  participant DB as PostgreSQL
  participant MQ as RabbitMQ

  loop every OutboxPollInterval
    Outbox->>Repo: ListPendingOutbox(now, staleBefore, batchSize)
    Repo->>DB: ClaimPendingChatEventOutbox(retryBefore=now, staleBefore)
    alt pending event reached retry time
      DB->>DB: status pending -> publishing
    else publishing event became stale
      DB->>DB: reclaim stale publishing row
    end
    DB-->>Repo: claimed publishing events
    Repo-->>Outbox: events

    loop each event
      Outbox->>MQ: Publish(exchange, routingKey, payload)
      alt publisher confirm ack
        MQ-->>Outbox: ack
        Outbox->>Repo: MarkOutboxPublished(id, now)
        Repo->>DB: status publishing -> published
      else publish error or timeout
        MQ-->>Outbox: error
        Outbox->>Outbox: calculate nextRetryAt
        Outbox->>Repo: MarkOutboxFailed(id, nextRetryAt, maxAttempts, now)
        alt attempts reached maxAttempts
          Repo->>DB: status publishing -> failed
        else can retry later
          Repo->>DB: status publishing -> pending
        end
      end
    end
  end
```

## RabbitMQ 消费侧 DLQ 重试链路

```mermaid
sequenceDiagram
  autonumber
  participant MQ as RabbitMQ chat.event
  participant MainQ as instance main queue
  participant Consumer as Worker
  participant RetryQ as instance retry queue
  participant DeadQ as instance dead queue
  participant Hub as session.Hub

  MQ->>MainQ: route chat.direct/chat.group
  MainQ-->>Consumer: deliver message(attempt=1)
  Consumer->>Consumer: handle event
  alt handled
    Consumer->>Hub: deliver online message
    Consumer->>MainQ: ack
  else discard error
    Consumer->>MainQ: ack and drop
  else retryable error and attempt < ConsumerMaxAttempts
    Consumer->>MainQ: nack requeue=false
    MainQ->>RetryQ: dead-letter by original routing key
    RetryQ->>RetryQ: wait ConsumerRetryDelay
    RetryQ->>MainQ: dead-letter to default exchange by queue name
    MainQ-->>Consumer: deliver message(attempt+1)
  else retryable error and attempt >= ConsumerMaxAttempts
    Consumer->>DeadQ: publish failed message with x-error
    Consumer->>MainQ: ack
  end
```

## 会话同步与已读链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as ChatHandler
  participant Logic as ChatLogic
  participant Repo as ChatRepo
  participant DB as PostgreSQL

  Client->>Handler: GET /api/chat/direct/:user_id/messages?before=&after=
  Handler->>Logic: ListDirectMessages(ctx, currentUserID, partnerID, query)
  Logic->>Logic: validate cursor and before/after exclusivity
  Logic->>Repo: ListDirectMessagesLatest/Before/After
  Repo->>DB: query chat_direct_message by both directions
  DB-->>Repo: direct messages
  Repo-->>Logic: message rows
  Logic-->>Handler: ChatDirectMessageListResp
  Handler-->>Client: response

  Client->>Handler: POST /api/chat/direct/:user_id/seen
  Handler->>Logic: MarkDirectSeen(ctx, currentUserID, partnerID, 0)
  Logic->>Repo: MarkDirectSeen(currentUserID, partnerID, now, now)
  Repo->>DB: upsert chat_direct_seen
  DB-->>Repo: ok
  Repo-->>Logic: nil
  Logic-->>Handler: nil
  Handler-->>Client: response

  Client->>Handler: GET /api/chat/groups/:group_id/messages?before=&after=
  Handler->>Logic: ListMessages(ctx, currentUserID, groupID, query)
  Logic->>Repo: GetMemberRole(groupID, currentUserID)
  Repo->>DB: select chat_group_member.role
  DB-->>Repo: member role
  Logic->>Repo: ListMessagesLatest/Before/After
  Repo->>DB: query chat_group_message
  DB-->>Repo: group messages
  Repo-->>Logic: message rows
  Logic-->>Handler: ChatGroupMessageListResp
  Handler-->>Client: response

  Client->>Handler: POST /api/chat/groups/:group_id/seen
  Handler->>Logic: MarkGroupSeen(ctx, currentUserID, groupID, 0)
  Logic->>Repo: UpdateMemberLastSeenTime(groupID, currentUserID, now)
  Repo->>DB: update chat_group_member.last_seen_time
  DB-->>Repo: ok
  Repo-->>Logic: nil
  Logic-->>Handler: nil
  Handler-->>Client: response
```

## 群成员与订阅链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as ChatHandler
  participant Logic as ChatLogic
  participant Repo as ChatRepo
  participant DB as PostgreSQL
  participant Hub as session.Hub

  Client->>Handler: POST /api/chat/groups
  Handler->>Logic: CreateGroup(ctx, ownerID, req)
  Logic->>Repo: CreateGroup(groupID, ownerID, profile, limit, now)
  Repo->>DB: insert chat_group and owner chat_group_member
  DB-->>Repo: ok
  Repo-->>Logic: nil
  Logic-->>Handler: CreateChatGroupResp
  Handler-->>Client: response

  Client->>Handler: POST /api/chat/groups/:group_id/join
  Handler->>Logic: JoinGroup(ctx, userID, groupID)
  Logic->>Repo: JoinGroup(groupID, userID, now)
  Repo->>DB: lock chat_group
  Repo->>DB: insert chat_group_member and increment member_count
  DB-->>Repo: ok
  Repo-->>Logic: nil
  Logic-->>Handler: nil
  Handler-->>Client: response

  Client->>Handler: WS /ws/group send {"op":"subscribe","group_id"}
  Handler->>Logic: HandleGroupMessage(ctx, userID, message)
  Logic->>Logic: handleGroupSubscribe
  Logic->>Repo: GetMemberRole(groupID, userID)
  Repo->>DB: select chat_group_member.role
  DB-->>Repo: member role
  Repo-->>Logic: role
  Logic-->>Handler: Subscribe groupID and ack
  Handler->>Hub: Subscribe(client, groupID)
  Handler-->>Client: ack

  Client->>Handler: WS /ws/group send {"op":"unsubscribe","group_id"}
  Handler->>Logic: HandleGroupMessage(ctx, userID, message)
  Logic-->>Handler: Unsubscribe groupID
  Handler->>Hub: Unsubscribe(client, groupID)
```
