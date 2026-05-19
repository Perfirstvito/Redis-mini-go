# Godis 手撕学习路线

> 学习周期：10天
> 目标：从零开始手写实现一个 Redis，1-2周可运行版本
> 每日完成：当天代码需能编译并运行测试

---

## 项目概览

Godis 是 Go 实现的 Redis 兼容服务器，支持单机和集群模式。

### 核心模块关系

```
main.go
├── config.SetupConfig() → config.Properties
├── if config.UseGnet
│   └── gnet.NewGnetServer(db)
│       └── db.Exec(conn, cmdLine)
└── else
    └── stdserver.MakeHandler()
        └── tcp.ListenAndServeWithSignal()
            └── handler.Handle()

db = cluster.MakeCluster()  OR  database.NewStandaloneServer()
```

### 模块速查

| 模块 | 路径 | 作用 |
|------|------|------|
| 配置 | `config/config.go` | 解析 redis.conf |
| 网络层 | `tcp/server.go` | TCP服务器，连接池 |
| 协议 | `redis/protocol/reply.go` | RESP编码 |
| 解析器 | `redis/parser/parser.go` | RESP解码 |
| 连接 | `redis/connection/conn.go` | 客户端状态 |
| 数据库 | `database/database.go` | 存储引擎 |
| 数据结构 | `datastruct/dict/concurrent.go` | 线程安全字典 |
| 命令 | `database/string.go` 等 | 各类型命令 |
| 持久化 | `aof/aof.go` | AOF/RDB |
| 集群 | `cluster/` | 分布式 |

---

## 学习路线总览

| 天数 | 主题 | 核心文件 | 验收标准 |
|------|------|----------|----------|
| **Day 1** | TCP服务器 + RESP协议 | `tcp/server.go`, `redis/protocol/reply.go`, `redis/parser/parser.go` | telnet发PING收到PONG |
| **Day 2** | 连接管理 | `redis/connection/conn.go` | 多客户端连接隔离 |
| **Day 3** | 基本数据结构 | `datastruct/dict/concurrent.go`, `datastruct/list/linked.go` | 字典+链表可并发使用 |
| **Day 4** | 数据库核心 | `database/database.go`, `database/server.go`, `database/router.go` | DBSIZE返回0 |
| **Day 5** | STRING命令 | `database/string.go`, `database/systemcmd.go` | SET/GET/INCR正常工作 |
| **Day 6** | LIST/HASH命令 | `database/list.go`, `database/hash.go` | LPUSH/LRANGE/HGET/HGETALL |
| **Day 7** | SET/SORTED SET | `database/set.go`, `database/sortedset.go` | SADD/SMEMBERS/ZADD/ZRANGE |
| **Day 8** | KEY命令 + 事务 | `database/keys.go`, `database/transaction.go` | EXISTS/EXPIRE/MULTI/EXEC |
| **Day 9** | 持久化 | `aof/aof.go`, `database/persistence.go` | SHUTDOWN后数据不丢失 |
| **Day 10** | 集群（可选） | `cluster/cluster.go`, `cluster/raft/raft.go` | 理解分布式概念 |

---

## 详细每日计划

---

## Day 1：TCP服务器 + Redis协议

**目标**: 建立网络层，理解RESP协议编码解码

### 上午：TCP服务器

**文件**:
```
tcp/server.go        → TCP服务器，带信号处理和优雅关闭
tcp/echo.go          → Echo服务器（测试用）
```

**重点理解**:
- `tcp/server.go:ListenAndServeWithSignal()` — 监听端口，接收连接
- 连接池管理：`map[string]net.Conn`
- 信号处理：`syscall.SIGINT`, `syscall.SIGTERM`

**手敲要点**:
- 简单实现：监听端口 → accept → 启动goroutine处理 → 循环
- 用 `bufio.NewReader` 和 `bufio.NewWriter` 包装连接
- Echo测试：用telnet连接，服务器原样返回输入

**下午：RESP协议

**文件**:
```
redis/protocol/consts.go    → 协议常量（\r\n、$、*、+、-、:）
redis/protocol/reply.go     → 回复类型结构体和编码方法
redis/protocol/errors.go   → 错误类型
redis/parser/parser.go     → RESP解析器
```

**RESP协议速记**:
- 简单字符串：`+OK\r\n`
- 错误：`-ERR\r\n`
- 整数：`:1\r\n`
- 批量字符串：`$3\r\nabc\r\n`
- 多批量字符串：`*2\r\n$3\r\nabc\r\n$3\r\ndef\r\n`

**手敲要点**:
- `reply.go`: 理解每种Reply的 `ToBytes()` 方法
- `parser.go`: 按字节解析，遇到 `\r\n` 分割

**验收测试**:
```bash
$ telnet localhost 6379
PING
+PONG
SET foo bar
+OK
GET foo
$3
bar
```

---

## Day 2：连接管理

**目标**: 管理客户端状态（读写缓冲、订阅、认证、事务）

### 文件

```
redis/connection/conn.go    → 连接结构体
redis/connection/fake.go   → 内部用假连接（如AOF加载时）
```

### 连接结构体字段

```go
type Conn struct {
    conn         net.Conn
    bufConn      *bufio.ReadWriter  // 带缓冲的读写
    waitingReply *reply.MultiBulkReply
    state        ConnState  // 事务状态
    subscriptions map[string]bool
    password     string
    selectedDB   int
}
```

### 重点理解

- `redis/connection/conn.go`: 连接的读写缓冲、订阅管理
- `state`: idle / in_transaction
- `selectedDB`: 当前选择的数据库编号（0-15）

**手敲要点**:
- 实现 `Write()` 方法：把Reply编码成bytes发送
- 实现 `ReadLine()` 方法：从bufio读取一行
- 实现 `GetDB()` / `SelectDB()` 方法

**验收测试**:
- 多个telnet连接同时连接，服务器独立响应
- SELECT 0 切换数据库

---

## Day 3：基本数据结构

**目标**: 实现核心数据结构——线程安全字典、双向链表、跳跃表

### 文件

```
datastruct/dict/dict.go         → Dict接口
datastruct/dict/simple.go      → 简单字典（非并发）
datastruct/dict/concurrent.go   → 并发字典（分片锁）
datastruct/list/interface.go   → List接口
datastruct/list/linked.go      → 双向链表
datastruct/list/quicklist.go  → QuickList
datastruct/sortedset/skiplist.go → 跳跃表
datastruct/sortedset/sortedset.go → SortedSet接口
```

### 并发字典（最重要）

```go
type ConcurrentDict struct {
    table  []*dict.SimpleDict   // 分片数组
    size   uint64
    mask   uint64
    locks  []sync.Mutex        // 每切片一个锁
}
```

**手敲要点**:
- FNV哈希计算槽位：`hash.FNV64a(key) % mask`
- 读写分离：读不加锁（只读共享数据），写加锁
- `Find()`: 计算hash → 定位分片 → 在分片内操作
- `Put()`: 加锁 → 查找/插入 → 解锁

### 双向链表

```go
type LinkedList struct {
    head *listNode
    tail *listNode
    len  int
}
```

### 跳跃表（sorted set底层）

- 层数随机（1-64）
- 每层向前跳跃
- 实现 `ZADD`, `ZRANGE` 的基础

**验收测试**:
```go
d := NewConcurrentDict(16)
d.Put("k1", "v1")
d.Put("k2", "v2")
val, _ := d.Get("k1")  // val == "v1"
// 并发Put不应panic
```

---

## Day 4：数据库核心

**目标**: 理解数据库引擎框架，注册命令

### 文件

```
database/database.go   → DB结构体
database/server.go    → Server结构体（多DB、AOF、复制）
database/commandinfo.go → 命令元信息
database/router.go    → 命令路由表（cmdTable）
```

### DB结构体

```go
type DB struct {
    data       *dict.ConcurrentDict  // 键值存储
    ttlMap     *dict.ConcurrentDict  // 过期时间
    versionMap *dict.ConcurrentDict  // 乐观锁版本
    lock       *sync.RWMutex
}
```

### Server结构体

```go
type Server struct {
    dbSet      []*DB
    persister  *aof.Persister
    masterConn *connection.FakeConn
    // ...
}
```

### 命令注册表

```go
var cmdTable = map[string]*CommandInfo{
    "PING": { executor: ping },
    "SET":  { executor: set },
    "GET":  { executor: get },
    // ...
}
```

**手敲要点**:
- `database.go`: 实现 `DB.Exec()` 入口
- `router.go`: cmdTable 映射命令名到执行函数
- `commandinfo.go`: 命令属性（参数个数、是否带事务）

**验收测试**:
```bash
$ redis-cli
> DBSIZE
(integer) 0
> PING
PONG
```

---

## Day 5：STRING命令

**目标**: 实现第一个数据类型命令

### 文件

```
database/string.go     → SET, GET, SETNX, INCR, DECR, APPEND, STRLEN
database/systemcmd.go → PING, AUTH, SELECT, DBSIZE, FLUSHDB, INFO
```

### STRING命令速查

| 命令 | 行为 | 实现要点 |
|------|------|----------|
| `SET key value` | 设置键值 | 直接写入dict |
| `GET key` | 获取值 | 从dict读 |
| `SETNX key value` | 不存在才设置 | 需要判断key存在 |
| `INCR key` | 自增1 | 转int后+1，需要原子操作 |
| `DECR key` | 自减1 | 转int后-1 |
| `APPEND key value` | 追加 | GET后拼接 |
| `STRLEN key` | 长度 | len(string) |

**手敲要点**:
- `string.go` 开头定义辅助函数 `ToString()`, `ToInt64()`
- INCR/DECR 需要先GET，转int，加1，再SET回去
- 注意处理不存在的键：GET返回nil时视为0

**验收测试**:
```bash
> SET foo bar
> GET foo
"bar"
> INCR counter
(integer) 1
> INCR counter
(integer) 2
> DECR counter
(integer) 1
> STRLEN foo
(integer) 3
> SETNX newkey value
(integer) 1
> SETNX newkey another
(integer) 0
> GET newkey
"value"
```

---

## Day 6：LIST和HASH命令

**目标**: 实现列表和哈希数据类型

### 文件

```
database/list.go      → LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, LINDEX
database/hash.go     → HSET, HGET, HGETALL, HDEL, HLEN, HEXISTS, HINCRBY
```

### LIST命令速查

| 命令 | 行为 |
|------|------|
| `LPUSH key value [value...]` | 头部插入，返回列表长度 |
| `RPUSH key value [value...]` | 尾部插入 |
| `LPOP key` | 头部弹出 |
| `RPOP key` | 尾部弹出 |
| `LLEN key` | 列表长度 |
| `LRANGE key start stop` | 范围查询 |
| `LINDEX key index` | 按索引访问 |

### HASH命令速查

| 命令 | 行为 |
|------|------|
| `HSET key field value` | 设置哈希字段 |
| `HGET key field` | 获取字段值 |
| `HGETALL key` | 获取所有字段值 |
| `HDEL key field [field...]` | 删除字段 |
| `HLEN key` | 字段数量 |
| `HEXISTS key field` | 字段是否存在 |
| `HINCRBY key field increment` | 字段值增加 |

**手敲要点**:
- LIST用 `datastruct/list/linked.go` 的双向链表
- HASH在DB.data中嵌套存储：`key -> map[field]value`
- LRANGE负索引转换：`-1` → `len-1`

**验收测试**:
```bash
> RPUSH mylist a b c
(integer) 3
> LRANGE mylist 0 -1
1) "a"
2) "b"
3) "c"
> LPOP mylist
"a"
> HSET user name Alice age 30
(integer) 2
> HGET user name
"Alice"
> HGETALL user
1) "name"
2) "Alice"
3) "age"
4) "30"
> HINCRBY user age 1
(integer) 31
```

---

## Day 7：SET和SORTED SET

**目标**: 实现集合和有序集合

### 文件

```
database/set.go         → SADD, SREM, SMEMBERS, SISMEMBER, SCARD, SPOP
database/sortedset.go   → ZADD, ZRANGE, ZREVRANGE, ZSCORE, ZRANK, ZREM, ZCARD
```

### SET命令速查

| 命令 | 行为 |
|------|------|
| `SADD key member [member...]` | 添加元素，返回添加数量 |
| `SREM key member [member...]` | 移除元素 |
| `SMEMBERS key` | 获取所有元素 |
| `SISMEMBER key member` | 是否存在 |
| `SCARD key` | 集合大小 |
| `SPOP key` | 随机弹出一个 |

### SORTED SET命令速查

| 命令 | 行为 |
|------|------|
| `ZADD key score member [score member...]` | 添加，score排序 |
| `ZRANGE key start stop [WITHSCORES]` | 按分数升序 |
| `ZREVRANGE key start stop [WITHSCORES]` | 降序 |
| `ZSCORE key member` | 获取分数 |
| `ZRANK key member` | 获取排名 |
| `ZREM key member` | 删除 |
| `ZINCRBY key increment member` | 增加分数 |

**手敲要点**:
- SET用 `datastruct/set/set.go` 的hash set
- SORTED SET用 `datastruct/sortedset/skiplist.go` 跳跃表
- 跳跃表节点：`score` + `member` + 前后向指针数组

**验收测试**:
```bash
> SADD fruits apple banana cherry
(integer) 3
> SISMEMBER fruits apple
(integer) 1
> SISMEMBER fruits grape
(integer) 0
> SMEMBERS fruits
> ZADD leaderboard 100 player1
(integer) 1
> ZADD leaderboard 200 player2
(integer) 1
> ZADD leaderboard 150 player3
(integer) 1
> ZRANGE leaderboard 0 -1 WITHSCORES
1) "player1"
2) "100"
3) "player3"
4) "150"
5) "player2"
6) "200"
> ZSCORE leaderboard player2
"200"
```

---

## Day 8：KEY命令 + 事务

**目标**: 键管理和MULTI/EXEC事务

### 文件

```
database/keys.go        → DEL, EXISTS, EXPIRE, TTL, RENAME, TYPE, KEYS, MOVE
database/transaction.go → MULTI, EXEC, DISCARD, WATCH
database/tx_utils.go
```

### KEY命令速查

| 命令 | 行为 |
|------|------|
| `DEL key [key...]` | 删除键 |
| `EXISTS key [key...]` | 存在返回1/0 |
| `EXPIRE key seconds` | 设置过期时间 |
| `TTL key` | 剩余生存时间，-1永不过期，-2不存在 |
| `RENAME key newkey` | 重命名 |
| `TYPE key` | 类型（string/list/hash/set/zset） |
| `KEYS pattern` | 模式匹配（简单支持*） |

### 事务命令速查

| 命令 | 行为 |
|------|------|
| `MULTI` | 开始事务 |
| `EXEC` | 执行队列中所有命令 |
| `DISCARD` | 清空队列 |
| `WATCH key [key...]` | 乐观锁，键变化则事务失败 |

**手敲要点**:
- EXPIRE利用 `lib/timewheel/timewheel.go` 的时间轮
- 事务状态：idle → in_transaction
- WATCH通过 `DB.versionMap` 实现：执行前比较版本号
- EXEC失败返回空（nil）

**验收测试**:
```bash
> SET key1 value1
> EXISTS key1
(integer) 1
> EXISTS notexist
(integer) 0
> EXPIRE key1 10
(integer) 1
> TTL key1
(integer) 10
> RENAME key1 key2
> TYPE key2
string
> MULTI
> SET a 1
> INCR a
> EXEC
> GET a
2
> WATCH mykey
> MULTI
> INCR mykey
> EXEC
(nil)  # 因为WATCH的键被修改
```

---

## Day 9：持久化

**目标**: 实现数据落盘（SHUTDOWN后数据不丢失）

### 文件

```
aof/aof.go            → AOF异步写入器
aof/marshal.go        → 命令序列化
aof/rewrite.go        → AOF重写
aof/rdb.go            → RDB格式支持
database/persistence.go → RDB保存
```

### AOF流程

```
Server.Exec(cmd)
    ↓
cmd executor（如SET foo bar）
    ↓
同时写入 persister
    ↓
persister接收命令，写入buffer
    ↓
后台goroutine循环flush到磁盘
```

### fsync策略

| 策略 | 行为 |
|------|------|
| `always` | 每次写操作后fsync |
| `everysec` | 每秒fsync一次 |
| `no` | 由操作系统决定 |

### RDB流程

```
SHUTDOWN或BGSAVE命令
    ↓
遍历所有DB
    ↓
将每个键值对写成RDB格式
    ↓
写入临时文件，rename替换
```

**手敲要点**:
- `aof/aof.go`: 启动后台goroutine，用channel或ring buffer通信
- `aof/marshal.go`: 把cmdLine转成字符串存储
- `persistence.go`: 使用 `gob` 或手写二进制格式

**验收测试**:
```bash
> SET persistent_key "survives_restart"
> SHUTDOWN NOSAVE
# 重启服务器
> GET persistent_key
"survives_restart"
```

---

## Day 10：集群基础（可选）

**目标**: 理解分布式概念

### 文件

```
cluster/cluster.go    → 集群入口
cluster/core/core.go  → 集群核心（槽分配）
cluster/core/node_manager.go → 节点管理
cluster/raft/raft.go → Raft共识封装
cluster/raft/fsm.go   → 状态机
```

### 集群核心概念

- **槽（slot）**: 16384个槽，key通过CRC16取模映射到槽
- **节点**: 每个节点负责一部分槽
- **Raft**: 写入操作通过Raft共识同步到所有节点
- **TCC**: Try-Confirm-Cancel分布式事务

**学习要点**:
- `cluster/core/core.go`: 槽分配表 `slotsManager`
- `cluster/raft/raft.go`: 封装 `hashicorp/raft` 库
- `raft/fsm.go`: 日志应用后修改slot ownership

**验收标准**:
- 理解key → slot的映射过程
- 理解Raft的Propose和Apply流程

---

## 每日验收脚本

### Day 1
```bash
# 用nc或telnet测试
$ nc localhost 6379
PING
+PONG
QUIT
+OK
```

### Day 4
```bash
$ redis-cli
> PING
PONG
> DBSIZE
(integer) 0
```

### Day 5
```bash
> SET day5key day5value
> GET day5key
> INCR day5counter
> INCR day5counter
> GET day5counter
```

### Day 6
```bash
> RPUSH day6list a b c
> LRANGE day6list 0 -1
> HSET day6hash name alice
> HGET day6hash name
> HGETALL day6hash
```

### Day 7
```bash
> SADD day7set x y z
> SMEMBERS day7set
> ZADD day7zset 10 a 20 b 30 c
> ZRANGE day7zset 0 -1 WITHSCORES
```

### Day 8
```bash
> SET day8key value
> EXISTS day8key
> EXPIRE day8key 3600
> TTL day8key
> MULTI
> SET t1 1
> INCR t1
> EXEC
```

### Day 9
```bash
# 配置开启AOF
> SET before_shutdown "data"
# 修改配置启用AOF，或执行BGSAVE
> SHUTDOWN
# 重启后
> GET before_shutdown
```

---

## 学习建议

1. **按顺序来**: 不要跳步，Day1-4是基础，后面命令依赖前面的数据结构
2. **每行手敲**: 不要复制粘贴，自己理解每个变量含义
3. **每天运行**: 每天保证代码能 `go build ./...` 通过
4. **参考Redis文档**: 对照官方文档理解每个命令的具体行为
5. **测试驱动**: 每实现一个命令就用 redis-cli 或 telnet 验证

---

## 快速参考

### 项目结构

```
godis/
├── main.go                    # 入口
├── config/config.go           # 配置
├── tcp/server.go              # TCP服务器
├── redis/
│   ├── protocol/reply.go      # 协议
│   ├── parser/parser.go       # 解析器
│   └── connection/conn.go      # 连接
├── database/
│   ├── server.go              # 数据库服务器
│   ├── database.go            # 单个数据库
│   ├── router.go              # 命令路由
│   ├── string.go              # STRING命令
│   ├── list.go                # LIST命令
│   ├── hash.go                # HASH命令
│   ├── set.go                 # SET命令
│   ├── sortedset.go           # SORTED SET命令
│   ├── keys.go                # KEY命令
│   └── transaction.go          # 事务
├── datastruct/
│   ├── dict/concurrent.go     # 并发字典
│   ├── list/linked.go         # 链表
│   └── sortedset/skiplist.go   # 跳跃表
├── aof/aof.go                 # AOF持久化
└── cluster/                   # 集群（可选）
```

### 数据类型存储方式

| Redis类型 | Go实现 |
|-----------|--------|
| STRING | 直接存储 `string` |
| LIST | `LinkedList` |
| HASH | `map[string]string` 在dict中 |
| SET | `Set` (hash表实现) |
| ZSET | `SkipList` |
| KEY过期 | `TimeWheel` 或 TTL map + 定时清理 |

### RESP协议速查

```
+OK                    简单字符串
-ERR                  错误
:100                  整数
$3\r\nabc\r\n         批量字符串（$长度\r\n内容\r\n）
*2\r\n$3\r\nabc\r\n   多批量字符串（*数量\r\n）
```