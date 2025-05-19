package matcher

import (
	"errors"
	"fmt"
	"strings"
)

type KeyMatcher = func(sub string) (string, bool)
type PathSpliter = func(string) ([]string, error)

type Matcher[T any] struct {
	matchParam    KeyMatcher
	matchWildcard KeyMatcher
	split         PathSpliter
	root          *pathTrieNode[T]
}

var (
	ErrInvalidPath = errors.New("invalid path")
)

var (
	RouterParamMatcher KeyMatcher = func(s string) (string, bool) {
		if len(s) <= 1 {
			return s, false
		}
		if s[0:1] == ":" {
			return s[1:], true
		} else {
			return s, false
		}
	}
	RouterWildcardMatcher KeyMatcher = func(sub string) (string, bool) {
		if len(sub) == 0 {
			return sub, false
		}
		if sub[0:1] == "*" {
			return "*", true
		} else {
			return sub, false
		}
	}
)

var (
	MqttTopicParamMatcher KeyMatcher = func(sub string) (string, bool) {
		if sub == "+" {
			return "+", true
		}
		return sub, false
	}
	MqttTopicWildMatcher KeyMatcher = func(sub string) (string, bool) {
		return sub, sub == "#"
	}

	MqttTopicPathSpliter PathSpliter = func(s string) ([]string, error) {
		if len(s) <= 0 {
			return nil, ErrInvalidPath
		}
		return strings.Split(s, "/"), nil
	}

	NatsSubjectParamMatcher KeyMatcher = func(sub string) (string, bool) {
		if sub == ">" {
			return ">", true
		}
		return sub, false
	}

	NatsSubjectWildMatcher KeyMatcher = func(sub string) (string, bool) {
		if sub == "*" {
			return "*", true
		}
		return sub, false
	}

	NatsSubjectPathSpliter PathSpliter = func(s string) ([]string, error) {
		if len(s) <= 0 {
			return nil, ErrInvalidPath
		}
		return strings.Split(s, "."), nil
	}
)

func NewMatcher[T any](paramMatcher, wildcardMatcher KeyMatcher, spliter PathSpliter) Matcher[T] {
	return Matcher[T]{
		matchParam:    paramMatcher,
		matchWildcard: wildcardMatcher,
		root: &pathTrieNode[T]{
			subPath:  "/",
			fullPath: "",
		},
		split: spliter,
	}
}

func NewMqttTopicMatcher[T any]() Matcher[T] {
	return NewMatcher[T](MqttTopicParamMatcher, MqttTopicWildMatcher, MqttTopicPathSpliter)
}

func NewRouterPathMatcher[T any]() Matcher[T] {
	return NewMatcher[T](RouterParamMatcher, RouterWildcardMatcher, MqttTopicPathSpliter)
}

func NewNatsSubjectMatcher[T any]() Matcher[T] {
	return NewMatcher[T](NatsSubjectParamMatcher, NatsSubjectWildMatcher, NatsSubjectPathSpliter)
}

func (t *Matcher[T]) AddPath(path string) error {
	return t.AddPathWithPriority(path, 0)
}

func (t *Matcher[T]) AddPathWithPriority(path string, prior int) error {
	var nilVal T
	return t.addSub(path, prior, nilVal)
}

func (t *Matcher[T]) AddPathWithValue(fullPath string, value T) error {
	return t.addSub(fullPath, 0, value)
}

func (t *Matcher[T]) addSub(path string, prior int, value T) error {
	subs, err := t.split(path)
	if err != nil {
		return err
	}
	node := t.root
	for _, sub := range subs {
		existed := false
		key, isParam, isWildcard := t.getPathKey(sub)
		for _, subNode := range node.arrayChilds {
			if key == subNode.subPath {
				node = subNode
				existed = true
				break
			}
		}
		if existed {
			continue
		}
		subNode := &pathTrieNode[T]{
			subPath:   key,
			fullPath:  "",
			wordFlag:  false,
			paramFlag: isParam,
			wildFlag:  isWildcard,
			priority:  prior,
		}
		if isWildcard {
			subNode.fullPath = path
		}
		node.arrayChilds = append(node.arrayChilds, subNode)
		sortArrayNodes(node.arrayChilds)
		node = subNode
	}
	node.wordFlag = true
	node.fullPath = path
	node.values = append(node.values, value)
	return nil
}

func (t *Matcher[T]) getPathKey(sub string) (key string, isParam, isWildcard bool) {
	key, isParam = t.matchParam(sub)
	if !isParam {
		key, isWildcard = t.matchWildcard(sub)
	}
	return
}

func (t *Matcher[T]) Match(dstTopic string) (matchedPath string, params map[string]string, ok bool) {
	subs, err := t.split(dstTopic)
	if err != nil {
		return "", nil, false
	}
	matchedPath, params, _, ok = t.root.match(subs)
	return
}

func (t *Matcher[T]) MatchWithValues(dstTopic string) (matchedPath string, params map[string]string, values []T, ok bool) {
	subs, err := t.split(dstTopic)
	if err != nil {
		return "", nil, nil, false
	}
	return t.root.match(subs)
}

func (t *Matcher[T]) MatchAll(dstTopic string) []MatchedResult[T] {
	subs, err := t.split(dstTopic)
	if err != nil {
		return nil
	}
	return t.root.matchAll(subs)
}

func (t *Matcher[T]) MatchWithAnonymousParams(dstTopic string) (matchedPath string, params []string, ok bool) {
	subs, err := t.split(dstTopic)
	if err != nil {
		return "", nil, false
	}
	matchedPath, params, _, ok = t.root.matchWithAnonymousParamsByTail(subs)
	return
}

func (t *Matcher[T]) MatchWithAnonymousParamsAndValues(dstTopic string) (matchedPath string, params []string, values []T, ok bool) {
	subs, err := t.split(dstTopic)
	if err != nil {
		return "", nil, nil, false
	}
	return t.root.matchWithAnonymousParamsByTail(subs)
}

func (t *Matcher[T]) Delete(path string) {
	subs, err := t.split(path)
	if err != nil {
		return
	}
	t.root.delete(subs, 0)
}

func (m *Matcher[T]) PrintTree() {
	fmt.Println("Prefix Trie Structure:")
	m.printNode(m.root, 0)
}

func (m *Matcher[T]) printNode(node *pathTrieNode[T], depth int) {
	// 缩进格式
	prefix := strings.Repeat("  ", depth)
	// 构建标志描述
	flags := ""
	if node.paramFlag {
		flags += "[Param]"
	}
	if node.wildFlag {
		flags += "[Wildcard]"
	}
	if node.wordFlag {
		flags += "[End]"
	}
	// 打印节点信息
	fmt.Printf(
		"%s- %s: fullPath=%q, flags=%s, priority=%d, values=%d\n",
		prefix, node.subPath, node.fullPath, flags, node.priority, len(node.values),
	)
	// 递归打印子节点（按优先级排序）
	for _, child := range sortArrayNodes(node.arrayChilds) {
		m.printNode(child, depth+1)
	}
}
