# Path Matcher GoDoc
高性能路径匹配库，支持多种协议模式匹配与参数捕获，适用于构建路由器、消息代理等系统组件。

## 📌 特性
* 多模式匹配：支持精确匹配、参数捕获(:param)、通配符(*)
* 多值存储：单路径可关联多个值
* 优先级排序：自动处理路径冲突（精确 > 参数 > 通配符）
* 协议适配：
    * HTTP路由风格
    * MQTT主题匹配
    * NATS主题匹配
* 高性能：优化的Trie树结构 + 尾递归优化

## 📦 安装
```bash
go get github.com/yourusername/matcher
```

## 🚀 快速开始
```go
import "github.com/yourusername/matcher"

// 创建HTTP路由匹配器
matcher := matcher.NewRouterPathMatcher[string]()

// 添加带值的路径
matcher.AddPathWithValue("/users/:id", "user_handler")

// 执行匹配
path, params, values, ok := matcher.MatchWithAnonymousParamsAndValues("/users/123")
// 输出: "/users/123", ["123"], ["user_handler"], true
```

## 🧩 支持的匹配模式
| 模式类型 |示例路径	| 匹配示例 |
| -------- | -------- | -------- |
| 精确匹配	| /home	 | /home |
| 参数捕获 |	/:year/:month |	/2023/04 |
| 通配符	| /api/* | 	/api/v1/resource |
| 混合模式	| /:version/api/*	| /v1/api/users/123|

## 🛠️ API 文档
### 核心类型
```go
type KeyMatcher = func(sub string) (string, bool)
type PathSpliter = func(string) ([]string, error)
```

### 创建匹配器
```go
func NewRouterPathMatcher[T any]() Matcher[T] // HTTP路由风格
func NewMqttTopicMatcher[T any]() Matcher[T]  // MQTT主题匹配
func NewNatsSubjectMatcher[T any]() Matcher[T] // NATS主题匹配
```

### 路径操作
```go
func (m *Matcher[T]) AddPath(path string) error
func (m *Matcher[T]) AddPathWithPriority(path string, prior int) error
func (m *Matcher[T]) AddPathWithValue(fullPath string, value T) error
```

### 匹配方法
```go
func (m *Matcher[T]) Match(dst string) (path string, params map[string]string, ok bool)
func (m *Matcher[T]) MatchWithValues(dst string) (path string, params map[string]string, values []T, ok bool)
func (m *Matcher[T]) MatchWithAnonymousParams(dst string) (path string, params []string, ok bool)
func (m *Matcher[T]) MatchWithAnonymousParamsAndValues(dst string) (path string, params []string, values []T, ok bool)
func (m *Matcher[T]) MatchAll(dst string) []MatchedResult[T]
```

## 📈 性能优化
* 内存优化：切片预分配 + 对象池复用
* 算法优化：子节点优先级排序减少回溯
* 并发安全：只读操作无锁设计

## 🧪 测试覆盖率
```bash
go test -cover ./...
```

## 🤝 贡献指南
1. Fork 仓库
2. 创建新分支 (git checkout -b feature/new)
3. 提交代码 (git commit -am 'Add new feature')
4. 推送分支 (git push origin feature/new)
5. 创建 Pull Request

## 📄 协议
MIT License. See LICENSE for full text.