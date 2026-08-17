# 馆际图书借阅服务

基于 Go 标准库的馆际图书借阅管理系统，支持多馆图书、副本、读者、借阅申请和运输交接。

## 核心概念

- **Book**: 书目信息
- **Copy**: 图书馆中的实体副本，同一副本同一时间只能服务一个借阅申请
- **Reader**: 读者，逾期未归还的读者不能发起新借阅
- **BorrowRequest**: 借阅申请，状态流转：`APPLIED` -> `LOCKED` -> `SHIPPED` -> `RECEIVED` -> `RETURNED`

## 功能

- 创建书目和副本
- 注册读者
- 发起借阅申请（自动校验读者逾期和副本是否可用）
- 锁定副本（仅当申请处于 APPLIED 状态）
- 发出副本（仅当申请处于 LOCKED 状态）
- 签收副本（仅当申请处于 SHIPPED 状态，并设置到期时间）
- 归还副本（仅当申请处于 RECEIVED 状态）
- 查询在途与逾期清单：列出所有 `SHIPPED` 和 `RECEIVED` 状态的申请，按借出馆和到期时间排序

## 快速开始

```bash
go build ./...
go test ./...
go run ./cmd/server
```

## API 示例

- `POST /readers` 注册读者
- `POST /books` 创建书目
- `POST /copies` 添加副本
- `POST /requests` 发起申请
- `POST /requests/{id}/lock` 锁定副本
- `POST /requests/{id}/ship` 发出副本
- `POST /requests/{id}/receive` 签收副本
- `POST /requests/{id}/return` 归还副本
- `GET /active-requests` 获取在途与逾期清单

## 规则与约束

- 副本状态必须为 `AVAILABLE` 才能被锁定
- 同一申请状态迁移必须严格按顺序
- 读者若有 `RECEIVED` 状态且已经逾期（当前日期大于到期时间）的申请，则不能创建新申请
- 在途清单按借出馆（图书馆名称）和到期时间排序
