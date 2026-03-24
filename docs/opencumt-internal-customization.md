# OpenCUMT 内部使用改造说明

本文档记录本次针对开源协会内部使用场景所做的定制化修改，覆盖邮箱域名限制与 Token 捐赠两个核心需求，并补充相关接口、数据结构、前端入口、数据库迁移以及当前验证情况。

适用时间：2026-03-24  
适用代码基线：当前工作区中的未提交改动

## 1. 改造目标

本次改动的目标有两项：

1. 将注册、邮箱验证码发送、邮箱绑定、OAuth 新建账号时使用的邮箱统一限制为 `@opencumt.org`。
2. 新增一个独立的 Token 捐赠功能，允许普通用户提交自己的上游 Token，由管理员审核后自动加入后台调用池。

## 2. 改动总览

本次修改分为前后端两部分：

- 后端新增统一邮箱校验工具，并在注册、邮箱绑定、邮箱验证码发送、OAuth 新建用户流程中接入。
- 前端新增 `@opencumt.org` 正则校验和输入归一化逻辑，在注册页与个人设置页提前拦截非法邮箱。
- 后端新增 `TokenDonation` 数据模型、用户提交接口、管理员审核接口、渠道自动创建逻辑。
- 前端新增 Token 捐赠页面、菜单入口和路由，用于普通用户提交捐赠、查看自己的记录，以及管理员审核。
- 数据库迁移中新增 `token_donations` 表。

## 3. 邮箱域名限制改造

### 3.1 后端改造

后端新增统一工具文件：

- `common/email_restriction.go`

该文件提供了以下能力：

- `NormalizeEmail(email string) string`
  - 统一做 `trim + lowercase` 归一化。
- `IsRestrictedRegisterEmail(email string) bool`
  - 使用正则校验邮箱是否匹配 `@opencumt.org`。
- `ValidateRestrictedRegisterEmail(email string) error`
  - 对外提供统一校验入口，失败时返回错误 `only @opencumt.org emails are allowed`。

校验使用的正则为：

```text
^[A-Za-z0-9._%+\-]+@opencumt\.org$
```

说明：

- 大小写不敏感，因为在匹配前会先做小写归一化。
- 会先去掉首尾空格，避免用户因为复制粘贴导致校验失败。

### 3.2 后端接入位置

#### 1. 注册流程

文件：

- `controller/user.go`

修改点：

- 用户注册时会先对邮箱做 `NormalizeEmail`。
- 然后执行 `ValidateRestrictedRegisterEmail`。
- 非 `@opencumt.org` 邮箱会被直接拒绝。
- 邮箱最终写入用户对象时使用归一化后的值，避免出现大小写不一致。

影响接口：

```text
POST /api/user/register
```

#### 2. 邮箱验证码发送

文件：

- `controller/misc.go`

修改点：

- 发送邮箱验证码前先对邮箱做归一化。
- 发送验证码接口只允许 `@opencumt.org` 邮箱。

影响接口：

```text
GET /api/verification?email=...
```

#### 3. 邮箱绑定

文件：

- `controller/user.go`

修改点：

- 绑定邮箱时会先归一化邮箱。
- 然后校验邮箱后缀是否为 `@opencumt.org`。
- 非协会邮箱无法完成绑定。

影响接口：

```text
GET /api/oauth/email/bind?email=...&code=...
```

#### 4. OAuth 新建账号

文件：

- `controller/oauth.go`

修改点：

- 当 OAuth 登录流程需要创建新账号时，会对 OAuth 返回的邮箱进行归一化和域名限制校验。
- 如果邮箱不是 `@opencumt.org`，将不允许自动创建新用户。

这意味着：

- OAuth 作为“新用户注册入口”时，同样受协会邮箱限制。
- 不会出现网页注册被限制、但 OAuth 可以绕过的情况。

#### 5. 密码找回相关处理

文件：

- `controller/misc.go`

修改点：

- 找回密码相关流程对邮箱也做了归一化处理。
- 当前没有额外把密码找回强制限制为 `@opencumt.org`，主要是为了避免历史已有账号在迁移阶段无法找回密码。

这部分是有意保守处理，不是遗漏。

### 3.3 JSON 处理顺手对齐项目规范

根据项目约定，业务代码中的 JSON 编解码应走 `common/json.go` 包装函数。

因此在本次邮箱限制相关改动中，顺手对以下文件做了对齐：

- `controller/user.go`
- `controller/misc.go`

主要替换为：

- `common.DecodeJson`
- `common.Marshal`
- `common.Unmarshal`

这样可以避免新增逻辑继续直接依赖 `encoding/json`。

### 3.4 前端改造

前端新增统一邮箱工具：

- `web/src/helpers/email.js`

提供以下内容：

- `OPENCUMT_EMAIL_HINT = '@opencumt.org'`
- `OPENCUMT_EMAIL_REGEX`
- `normalizeOrgEmail(value)`
- `isOpenCUMTEmail(value)`

并在：

- `web/src/helpers/index.js`

中统一导出，供各页面复用。

### 3.5 前端接入位置

#### 1. 注册页

文件：

- `web/src/components/auth/RegisterForm.jsx`

改动内容：

- 用户输入邮箱时会先做 `trim + lowercase`。
- 发送邮箱验证码前会先校验是否为 `@opencumt.org`。
- 提交注册前会再次校验。
- 输入框提示文案已明确说明“仅支持使用 `@opencumt.org` 邮箱”。

行为效果：

- 用户在前端就会被提前拦截。
- 即使前端被绕过，后端也会再次校验。

#### 2. 个人设置页邮箱绑定

文件：

- `web/src/components/settings/PersonalSetting.jsx`

改动内容：

- 邮箱输入时统一做归一化。
- 发送验证码前校验邮箱后缀。
- 执行邮箱绑定前再次校验邮箱后缀。

行为效果：

- 个人中心中绑定邮箱也被统一限制为 `@opencumt.org`。

### 3.6 邮箱限制的最终生效范围

当前已经覆盖以下场景：

- 用户名注册
- 邮箱验证码发送
- 邮箱绑定
- OAuth 新用户创建

当前未强制限制的场景：

- 密码找回

设计原因：

- 避免项目中历史上已经存在的非协会邮箱账号在迁移阶段无法进行密码重置。

### 3.7 相关文件

- `common/email_restriction.go`
- `common/email_restriction_test.go`
- `controller/misc.go`
- `controller/oauth.go`
- `controller/user.go`
- `web/src/helpers/email.js`
- `web/src/helpers/index.js`
- `web/src/components/auth/RegisterForm.jsx`
- `web/src/components/settings/PersonalSetting.jsx`

## 4. Token 捐赠功能

### 4.1 功能定位

新增的 Token 捐赠功能面向普通用户开放，目标是让普通用户可以将自己可用的上游 Token 提交给平台，由管理员审核后自动转化为后台可调用的渠道，从而加入调用池。

这是一个独立的新功能，原项目中没有这一套流程。

### 4.2 整体流程

完整流程如下：

1. 普通用户在前端填写捐赠表单。
2. 后端校验提交内容是否合法。
3. 后端把记录写入 `token_donations` 表，状态为 `pending`。
4. 管理员在后台查看待审核记录。
5. 如果管理员通过审核：
   - 后端自动创建一条 `channel` 记录。
   - 自动调用 `AddAbilities` 为该渠道写入模型能力。
   - 将捐赠状态更新为 `approved`。
   - 回填 `channel_id`，建立“捐赠记录 -> 实际渠道”的关联。
6. 如果管理员拒绝：
   - 将捐赠状态更新为 `rejected`。

### 4.3 数据模型

新增文件：

- `model/token_donation.go`

新增模型：

- `TokenDonation`

主要字段如下：

| 字段 | 含义 |
|------|------|
| `id` | 主键 |
| `user_id` | 提交捐赠的用户 ID |
| `type` | 渠道类型 |
| `name` | 渠道名称 |
| `key` | 用户提交的上游 Token / Key |
| `base_url` | 可选，自定义上游 Base URL |
| `models` | 模型列表，逗号分隔 |
| `group` | 渠道分组 |
| `remark` | 备注 |
| `openai_organization` | 可选，OpenAI / Azure 组织字段 |
| `status` | 审核状态 |
| `review_note` | 预留审核备注字段 |
| `channel_id` | 审核通过后生成的实际渠道 ID |
| `created_time` | 提交时间 |
| `reviewed_time` | 审核时间 |
| `reviewed_by` | 审核管理员 ID |

状态常量：

- `pending`
- `approved`
- `rejected`

### 4.4 数据访问层能力

`model/token_donation.go` 中提供了以下方法：

- `Insert()`
  - 新增捐赠记录。
- `Save()`
  - 保存更新后的捐赠记录。
- `GetTokenDonationById(id int)`
  - 按 ID 获取单条记录。
- `GetUserTokenDonations(userId int)`
  - 获取某个用户自己的捐赠记录。
- `GetTokenDonationsWithUsers(status string)`
  - 获取管理员视角的捐赠记录，并联表返回用户信息。

### 4.5 后端控制器

新增文件：

- `controller/token_donation.go`

核心能力包括：

#### 1. 支持的渠道类型限制

当前允许捐赠的渠道类型包括：

- OpenAI
- OpenAIMax
- Azure
- Anthropic
- Ali
- Tencent
- OpenRouter
- Gemini
- Moonshot
- Zhipu_v4
- Perplexity
- Cohere
- SiliconFlow
- Mistral
- DeepSeek
- VolcEngine
- xAI

没有加入白名单的渠道类型，提交时会被拒绝。

#### 2. 提交参数校验

后端会校验以下内容：

- `type` 必须属于允许捐赠的渠道类型。
- `name` 必填。
- `key` 必填。
- `models` 必填。
- `group` 为空时自动补成 `default`。
- `name` 长度不能超过 128。
- `remark` 长度不能超过 255。
- `openai_organization` 长度不能超过 255。
- `base_url` 如果填写，必须是合法的 `http` 或 `https` 地址。

此外还有两类归一化处理：

- `models` 和 `group` 会按逗号拆分、去空格、去重后重新拼接。
- 去重时按不区分大小写处理，避免重复模型名或分组名。

提交时还会调用项目现有的 `validateChannel(channel, true)`，确保转换成真正渠道之前，配置本身已符合原系统的渠道校验规则。

#### 3. 返回数据脱敏

接口返回捐赠数据时，不会直接回显完整 Key，而是使用现有的 `model.MaskTokenKey` 做脱敏。

#### 4. 审核通过后的渠道生成逻辑

管理员审核通过时，后端会基于捐赠记录构造一个新的 `model.Channel`：

- `Status` 直接设为启用状态。
- `Models` 和 `Group` 使用捐赠表单中的值。
- `BaseURL`、`OpenAIOrganization` 透传到渠道。
- `Tag` 自动设置为 `token-donation`。
- `Remark` 自动写成：

```text
[Donation #捐赠ID][User #用户ID] 原备注
```

并限制最长 255 字符。

之后在数据库事务中完成以下操作：

1. 再次确认该记录仍然处于 `pending`。
2. 创建实际渠道。
3. 调用 `channel.AddAbilities(tx)` 写入模型能力。
4. 更新捐赠记录状态为 `approved`。
5. 回写 `channel_id`、`reviewed_by`、`reviewed_time`。

这样可以保证“审核通过”与“渠道真正入池”是同一个事务内完成的，避免中间状态不一致。

#### 5. 拒绝逻辑

管理员拒绝时会更新：

- `status = rejected`
- `reviewed_by`
- `reviewed_time`

当前版本没有单独提供审核备注输入，虽然模型中保留了 `review_note` 字段，但接口和前端界面暂未使用。

### 4.6 新增接口

路由文件：

- `router/api-router.go`

新增接口如下。

#### 普通用户接口

```text
GET  /api/user/token_donation
POST /api/user/token_donation
```

说明：

- `GET /api/user/token_donation`
  - 获取当前登录用户自己的捐赠记录。
- `POST /api/user/token_donation`
  - 提交新的捐赠申请。

#### 管理员接口

```text
GET  /api/token_donation/?status=pending|approved|rejected|all
POST /api/token_donation/:id/approve
POST /api/token_donation/:id/reject
```

说明：

- `GET /api/token_donation/`
  - 管理员按状态筛选捐赠记录，并能看到提交用户的用户名和邮箱。
- `POST /api/token_donation/:id/approve`
  - 审核通过，自动创建渠道并加入调用池。
- `POST /api/token_donation/:id/reject`
  - 审核拒绝。

### 4.7 数据库迁移

文件：

- `model/main.go`

改动内容：

- 在 `migrateDB()` 中加入 `&TokenDonation{}`。
- 在 `migrateDBFast()` 中加入 `TokenDonation` 的迁移登记。

这意味着：

- 新部署实例会自动创建 `token_donations` 表。
- 已有实例在执行迁移时也会识别到该表。

### 4.8 前端页面与入口

新增页面：

- `web/src/pages/TokenDonation/index.jsx`

新增路由：

- `web/src/App.jsx`

新增侧边栏入口：

- `web/src/components/layout/SiderBar.jsx`
- `web/src/hooks/common/useSidebar.js`

页面布局适配：

- `web/src/components/layout/PageLayout.jsx`

#### 页面包含的功能

页面分为三个区域：

1. 提交捐赠
   - 普通用户填写渠道类型、名称、Token、模型列表、分组、Base URL、OpenAI Organization、备注等信息。
2. 我的捐赠记录
   - 普通用户查看自己提交过的记录、状态、入池后的渠道 ID、提交时间等。
3. 管理员审核
   - 管理员查看待审核或已审核记录，并执行通过 / 拒绝操作。

#### 前端表单字段

当前前端表单字段包括：

- `type`
- `name`
- `key`
- `base_url`
- `models`
- `group`
- `remark`
- `openai_organization`

#### 前端交互行为

- 用户提交后会提示“等待管理员审核”。
- 管理员通过后，前端会刷新列表，能看到记录已入池。
- 管理员可按 `pending`、`approved`、`rejected`、`all` 过滤查看。

### 4.9 Token 捐赠相关文件

- `controller/token_donation.go`
- `model/token_donation.go`
- `model/main.go`
- `router/api-router.go`
- `web/src/pages/TokenDonation/index.jsx`
- `web/src/App.jsx`
- `web/src/components/layout/SiderBar.jsx`
- `web/src/hooks/common/useSidebar.js`
- `web/src/components/layout/PageLayout.jsx`

## 5. 本次修改涉及的文件清单

### 后端

- `common/email_restriction.go`
- `common/email_restriction_test.go`
- `controller/misc.go`
- `controller/oauth.go`
- `controller/token_donation.go`
- `controller/user.go`
- `model/main.go`
- `model/token_donation.go`
- `router/api-router.go`

### 前端

- `web/src/App.jsx`
- `web/src/components/auth/RegisterForm.jsx`
- `web/src/components/layout/PageLayout.jsx`
- `web/src/components/layout/SiderBar.jsx`
- `web/src/components/settings/PersonalSetting.jsx`
- `web/src/helpers/email.js`
- `web/src/helpers/index.js`
- `web/src/hooks/common/useSidebar.js`
- `web/src/pages/TokenDonation/index.jsx`

## 6. 当前验证情况

已完成的验证：

- 新增了后端测试文件 `common/email_restriction_test.go`，用于校验邮箱归一化与 `@opencumt.org` 限制逻辑。
- 前端改动后的关键文件已做格式校验，语法可解析。

当前环境中的限制：

- 当前终端环境没有可用的 `go` / `gofmt`，因此无法在这里执行 Go 测试、Go 编译或 `gofmt`。
- 当前终端环境没有 `bun`，因此前端只能做 Node/npm 侧的辅助验证。

前端构建现状：

- `web/` 目录下执行打包时，仍然会遇到项目现有依赖问题：

```text
Missing "./dist/css/semi.css" specifier in "@douyinfe/semi-ui" package
```

这个问题属于项目当前依赖状态，不是本次改动新增的问题。

## 7. 当前版本的边界与后续可扩展点

当前已经满足需求，但仍有一些明确边界：

- Token 捐赠的“拒绝”动作暂时没有填写审核备注的前端交互。
- `review_note` 字段已经在模型里预留，后续如果需要可以很容易补充。
- 当前采用“管理员审核通过后自动入池”的模式，没有做“自动通过”或“先禁用后人工启用”的分层策略。
- 当前邮箱限制主要覆盖新注册与新绑定场景，对历史账号迁移策略未做额外批处理。

如果后续还需要继续迭代，比较自然的扩展方向包括：

- 给管理员审核页增加“审核备注”输入框，并写入 `review_note`。
- 给 Token 捐赠增加更细粒度的字段，例如有效期、来源说明、所有权说明、预计模型能力说明。
- 给捐赠渠道增加专用分组、优先级、权重、默认标签策略。
- 对历史非 `@opencumt.org` 用户做一次数据审计和迁移策略设计。

## 8. 结论

本次定制已经完成了两个核心目标：

1. 注册与绑定相关邮箱入口统一限制为 `@opencumt.org`，前后端双重校验。
2. 新增普通用户可提交、管理员可审核、审核通过后自动入池的 Token 捐赠功能。

从当前实现来看，这套改造已经能够支撑“协会内部账号体系 + 协会成员贡献上游 Token”的基本使用场景。
