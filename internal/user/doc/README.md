# User 模块主链路

本文档记录 `internal/user` 的主要业务链路。User 模块拥有账号、登录注册、联系方式绑定、伴侣关系、关注关系和用户管理的独立 SQL、repo/cache、handler/logic 边界。

## 模块启动链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant Module as user.Module
  participant Repo as user.repo
  participant Logic as user.logic
  participant Handler as user.handler
  participant Email as pkg/emailx
  participant SMS as pkg/smsx
  participant Baby as baby syncer

  Router->>Module: NewModule(DB, RDB, Log, Email?, SMS?, BabySyncer)
  Module->>Email: use injected Email or emailx.NewEmailX()
  Module->>SMS: use injected SMS or smsx.NewSmsX()
  Module->>Repo: NewUserRepo(DB, RDB, Log)
  Module->>Logic: NewUserLogic(repo, email, sms, BabySyncer, Log)
  Module->>Handler: NewUserHandler(logic)
  Router->>Module: RegisterRoutes(api.Group('/user'))
  Router->>Module: RegisterAdminRoutes(api.Group('/admin'))
  Logic-->>Baby: sync partner babies after partner binding
```

## 注册与登录链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as UserHandler
  participant Logic as UserLogic
  participant Email as email sender
  participant SMS as sms sender
  participant Repo as UserRepo
  participant DB as PostgreSQL
  participant JWT as pkg/jwtx

  Client->>Handler: POST /api/user/code/register
  Handler->>Logic: GetRegisterCode(ctx, req)
  Logic->>Email: SendCode(register key, email, code)
  Email-->>Logic: nil
  Logic-->>Handler: GetCodeResp

  Client->>Handler: POST /api/user/register
  Handler->>Logic: Register(ctx, req)
  Logic->>Email: VerifyCode(register key, code)
  Email-->>Logic: ok
  Logic->>Repo: Register(userID, username, email, account, password, gender)
  Repo->>DB: insert user with hashed password
  Repo-->>Logic: nil
  Logic-->>Handler: RegisterResp

  Client->>Handler: POST /api/user/login
  Handler->>Logic: Login(ctx, req)
  alt account password login
    Logic->>Repo: LoginWithAccount(account, password)
    Repo->>DB: query user by account and password hash
  else email code login
    Logic->>Email: VerifyCode(login key, code)
    Logic->>Repo: LoginWithEmail(email)
    Repo->>DB: query user by email
  end
  Repo-->>Logic: user row
  Logic->>JWT: GenToken(userID, role)
  JWT-->>Logic: token
  Logic-->>Handler: LoginResp
  Handler-->>Client: response
```

## 联系方式绑定链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as UserHandler
  participant Logic as UserLogic
  participant Sender as email or sms sender
  participant Repo as UserRepo
  participant DB as PostgreSQL

  Client->>Handler: POST /api/user/code/bind/phone
  Handler->>Logic: GetBindPhoneCode(ctx, userID, req)
  Logic->>Logic: validate phone format
  Logic->>Sender: SendCode(bind phone key, phone, code)
  Sender-->>Logic: nil
  Logic-->>Handler: GetCodeResp

  Client->>Handler: POST /api/user/bind/phone
  Handler->>Logic: BindPhone(ctx, userID, req)
  Logic->>Sender: VerifyCode(bind phone key, code)
  Sender-->>Logic: ok
  Logic->>Repo: BindPhone(userID, phone)
  Repo->>DB: update user phone
  Repo-->>Logic: nil
  Logic-->>Handler: BindContactResp
```

## 伴侣绑定链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as UserHandler
  participant Logic as UserLogic
  participant Repo as UserRepo
  participant DB as PostgreSQL
  participant Baby as baby syncer

  Client->>Handler: POST /api/user/partner/bind
  Handler->>Handler: auth user and bind PartnerBindReq
  Handler->>Logic: BindPartner(ctx, userID, req)
  Logic->>Repo: LoginWithAccount(partner account, partner password)
  Repo->>DB: verify partner credentials
  DB-->>Repo: partner user row
  Repo-->>Logic: partner row
  Logic->>Logic: validate gender and relation direction
  Logic->>Repo: BindPartner(fatherID, motherID)
  Repo->>DB: update both users partner fields
  Repo-->>Logic: nil
  Logic->>Baby: SyncPartnerBabies(fatherID, motherID)
  Baby-->>Logic: nil
  Logic-->>Handler: PartnerBindResp
  Handler-->>Client: response
```

## 关注链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as UserHandler
  participant Logic as UserLogic
  participant Repo as UserRepo
  participant DB as PostgreSQL
  participant Cache as user cache

  Client->>Handler: POST /api/user/follow/:target_user_id
  Handler->>Logic: Follow(ctx, userID, uri)
  Logic->>Logic: reject empty target or following self
  Logic->>Repo: FollowUser(followerID, followeeID)
  Repo->>DB: insert follow row
  Repo->>Cache: delete following/followers cache patterns
  Logic-->>Handler: FollowResp

  Client->>Handler: GET /api/user/following
  Handler->>Logic: ListFollowing(ctx, userID, req)
  Logic->>Repo: ListFollowing(viewID, page, pageSize)
  Repo->>Cache: read following list cache
  alt cache miss
    Repo->>DB: query following list
    Repo->>Cache: cache list page
  end
  Repo-->>Logic: rows and hasMore
  Logic-->>Handler: FollowingListResp
  Handler-->>Client: response
```

## 边界说明

- User 模块对外只暴露伴侣读取、关注读取等小接口给 router 注入到其它模块。
- Email/SMS 是基础设施能力，可以注入测试实现；默认使用 `internal/pkg/emailx` 和 `internal/pkg/smsx`。
- 伴侣绑定后的宝宝同步通过注入的 `BabySyncer` 完成，避免 user 模块直接依赖 baby 模块内部包。
