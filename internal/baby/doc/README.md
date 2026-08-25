# Baby 模块主链路

本文档记录 `internal/baby` 的主要业务链路。Baby 模块拥有宝宝资料、成长记录、疫苗记录、日常喂养/睡眠/尿布和相册的独立 SQL、repo/cache、handler/logic 边界。

## 模块启动链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant Module as baby.Module
  participant Repo as baby.repo
  participant Logic as baby.logic
  participant Handler as baby.handler
  participant BabyClient as baby.Client
  participant User as user.Client via PartnerReader
  participant Worker as baby.worker
  participant MQ as RabbitMQ

  Router->>Module: NewModule(DB, RDB, RabbitMQ, Log, PartnerReader)
  Module->>Repo: NewBabyRepo(DB, RDB, Log)
  Module->>Logic: NewBabyLogic(repo, PartnerReader, Log)
  Module->>Handler: NewBabyHandler(logic)
  Module->>BabyClient: NewClient(repo)
  alt DB and RabbitMQ enabled
    Module->>Worker: Start partner-bound consumer
    Worker->>MQ: consume user.event partner.bound
  end
  Router->>Module: RegisterRoutes(api.Group('/baby'))
  Router->>Module: RegisterAdminRoutes(api.Group('/admin'))
  Router-->>BabyClient: inject into AI boundary
  Logic-->>User: read partner relationship when creating baby
```

## Client 边界链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant BabyClient as baby.Client
  participant AI as ai.logic BabyGrowthReader
  participant Repo as BabyRepo

  Router->>BabyClient: babyModule.Client()
  Router->>AI: inject growth adapter backed by baby.Client
  AI->>BabyClient: GetBabyByIDAndUser(babyID, userID)
  BabyClient->>Repo: GetBabyByIDAndUser(babyID, userID)
  BabyClient-->>AI: baby.BabyProfile
  AI->>BabyClient: ListGrowthRecordsByBabyIDBetween(babyID, from, to)
  BabyClient->>Repo: ListGrowthRecordsByBabyIDBetween(babyID, from, to)
  BabyClient-->>AI: []baby.GrowthRecord
```

## 伴侣绑定事件消费链路

```mermaid
sequenceDiagram
  autonumber
  participant User as user outbox worker
  participant MQ as RabbitMQ
  participant Worker as baby.worker
  participant Logic as baby.event logic
  participant Repo as BabyRepo
  participant DB as PostgreSQL

  User->>MQ: publish user.event partner.bound
  MQ->>Worker: deliver event to durable baby.partner.bound queue
  Worker->>Worker: decode JSON and validate IDs
  Worker->>Logic: HandlePartnerBound(ctx, eventID, fatherID, motherID)
  Logic->>Repo: HandlePartnerBoundEvent(eventID, fatherID, motherID)
  Repo->>DB: begin transaction
  Repo->>DB: insert baby_event_inbox(eventID)
  alt inbox duplicate
    Repo-->>Logic: processed=false
    Logic-->>Worker: nil
    Worker-->>MQ: ack
  else first delivery
    Repo->>DB: copy father babies to mother
    Repo->>DB: copy mother babies to father
    Repo->>DB: mark inbox processed
    Repo->>DB: commit
    Repo-->>Logic: processed=true
    Logic-->>Worker: nil
    Worker-->>MQ: ack
  end
  alt decode or validation failed
    Worker-->>MQ: discard and ack
  else repo or database failed
    Worker-->>MQ: return error for retry or dead letter
  end
```

## 新建宝宝链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as BabyHandler
  participant Logic as BabyLogic
  participant User as user partner reader
  participant Repo as BabyRepo
  participant DB as PostgreSQL
  participant Cache as baby cache

  Client->>Handler: POST /api/baby
  Handler->>Handler: auth user and bind NewBabyReq
  Handler->>Logic: NewBaby(ctx, userID, req)
  Logic->>User: GetPartnerByUserID(userID)
  User-->>Logic: partnerID or empty
  Logic->>Repo: CreateBabyWithInit(userID, partnerID, baby profile, initial growth)
  Repo->>DB: begin transaction
  Repo->>DB: insert baby owner row
  alt initial growth exists
    Repo->>DB: insert baby_growth_record
  end
  Repo->>DB: list vaccine doses
  loop each dose
    Repo->>DB: insert baby_vaccine_record
  end
  alt partner exists
    Repo->>DB: insert partner baby row
    Repo->>DB: insert partner initial growth when present
  end
  Repo->>DB: commit
  Repo->>Cache: invalidate affected baby caches
  Repo-->>Logic: nil
  Logic-->>Handler: NewBabyResp
  Handler-->>Client: response
```

## 成长记录链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as BabyHandler
  participant Logic as BabyLogic
  participant Repo as BabyRepo
  participant DB as PostgreSQL
  participant Cache as baby cache

  Client->>Handler: POST /api/baby/growth-records
  Handler->>Logic: CreateGrowth(ctx, userID, req)
  Logic->>Repo: GetBabyByIDAndUser(babyID, userID)
  Repo->>DB: verify baby ownership
  DB-->>Repo: baby row
  Repo-->>Logic: ok
  Logic->>Repo: GetLatestGrowthByBabyID(babyID)
  Repo->>Cache: read latest growth
  alt same record date
    Logic->>Repo: UpdateGrowthByRecordID(recordID, metrics)
    Repo->>DB: update growth record
  else new date
    Logic->>Repo: CreateGrowthRecord(babyID, userID, metrics)
    Repo->>DB: insert growth record
  end
  Repo->>Cache: delete latest growth cache
  Logic-->>Handler: CreateGrowthResp
  Handler-->>Client: response
```

## 日常记录链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as BabyHandler
  participant Logic as BabyLogic
  participant Repo as BabyRepo
  participant DB as PostgreSQL

  Client->>Handler: POST /api/baby/:baby_id/daily/sleep/start
  Handler->>Logic: SleepStart(ctx, userID, uri)
  Logic->>Repo: GetBabyByIDAndUser(babyID, userID)
  Logic->>Repo: StartSleep(babyID, userID)
  Repo->>DB: close active state rules and insert sleep session
  Repo-->>Logic: sessionID, startedAt
  Logic-->>Handler: SleepStartResp

  Client->>Handler: POST /api/baby/:baby_id/daily/feeding
  Handler->>Logic: CreateFeeding(ctx, userID, uri, req)
  Logic->>Logic: validate feed type and time
  Logic->>Repo: CreateFeeding(babyID, userID, feed fields)
  Repo->>DB: insert feeding record
  Repo-->>Logic: feedingID
  Logic-->>Handler: FeedingCreateResp

  Client->>Handler: POST /api/baby/:baby_id/daily/diaper
  Handler->>Logic: CreateDiaper(ctx, userID, uri, req)
  Logic->>Logic: validate diaper fields and time
  Logic->>Repo: CreateDiaper(babyID, userID, diaper fields)
  Repo->>DB: insert diaper record
  Repo-->>Logic: diaperID
  Logic-->>Handler: DiaperCreateResp
```

## 疫苗与相册链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as BabyHandler
  participant Logic as BabyLogic
  participant Repo as BabyRepo
  participant DB as PostgreSQL
  participant Cache as baby cache

  Client->>Handler: GET /api/baby/vaccines
  Handler->>Logic: GetVaccineList(ctx, userID, req)
  Logic->>Repo: GetBabyByIDAndUser(babyID, userID)
  Logic->>Repo: ListVaccineRecordsByBaby(babyID)
  Repo->>Cache: read vaccine list cache
  alt cache miss
    Repo->>DB: query vaccine records and doses
    Repo->>Cache: cache vaccine list
  end
  Logic-->>Handler: GetVaccineListResp

  Client->>Handler: PUT /api/baby/vaccines/status
  Handler->>Logic: ChangeVaccineStatus(ctx, userID, req)
  Logic->>Repo: GetBabyByIDAndUser(babyID, userID)
  alt status is given
    Logic->>Repo: UpdateVaccineStatusGiven(babyID, doseID, actualTime)
  else status is not given
    Logic->>Repo: UpdateVaccineStatusNotGiven(babyID, doseID)
  end
  Repo->>DB: update vaccine record
  Repo->>Cache: delete vaccine list cache
  Logic-->>Handler: ChangeVaccineStatusResp

  Client->>Handler: POST /api/baby/photos
  Handler->>Logic: UploadBabyPhotos(ctx, userID, req)
  Logic->>Repo: GetBabyByIDAndUser(babyID, userID)
  Logic->>Repo: UploadPhotos(babyID, links)
  Repo->>DB: insert photo rows
  Logic-->>Handler: UploadBabyPhotosResp
```
