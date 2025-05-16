package matcher

import (
	"sort"
)

type priorities[T any] []*pathTrieNode[T]

func (p priorities[T]) Len() int {
	return len(p)
}

func (p priorities[T]) Swap(l, r int) {
	p[l], p[r] = p[r], p[l]
}

func (p priorities[T]) Less(l, r int) bool {
	return p[l].priority < p[r].priority
}

type pathTrieNode[T any] struct {
	subPath     string             //子路径
	fullPath    string             //完整路径
	arrayChilds []*pathTrieNode[T] //子树列表
	wordFlag    bool
	paramFlag   bool
	wildFlag    bool
	priority    int
	values      []T //用于存储路径值
}

func (t *pathTrieNode[T]) match(subs []string) (matchedPath string, params map[string]string, values []T, ok bool) {
	params = map[string]string{}
	matchedPath, ok = t.backtrace(t, subs, params, &values, 0)
	return
}

func (t *pathTrieNode[T]) matchWithAnonymousParamsByTail(subs []string) (matchedPath string, params []string, values []T, ok bool) {
	params = make([]string, 0)
	ok = t.backtraceWithTail(t, subs, &params, 0, &matchedPath, &values)
	return
}

type MatchedResult[T any] struct {
	Path   string
	Params []string
	Values []T
}

func (t *pathTrieNode[T]) backtrace(node *pathTrieNode[T], subs []string, params map[string]string, values *[]T, index int) (string, bool) {
	if index == len(subs) {
		*values = node.values
		return node.fullPath, true
	}
	nodes := node.arrayChilds
	for _, subNode := range nodes {
		if subNode.subPath == subs[index] {
			matched, ok := t.backtrace(subNode, subs, params, values, index+1)
			if ok {
				if index == len(subs)-1 {
					return subNode.fullPath, true
				} else {
					return matched, true
				}
			}
		} else if subNode.paramFlag {
			matched, ok := t.backtrace(subNode, subs, params, values, index+1)
			if ok {
				if index == len(subs)-1 {
					if subNode.wordFlag {
						params[subNode.subPath] = subs[index]
						return subNode.fullPath, true
					} else {
						continue
					}
				}
				params[subNode.subPath] = subs[index]
				return matched, true
			}
		} else if subNode.wildFlag {
			*values = subNode.values
			return subNode.fullPath, true
		} else {
			continue
		}
	}
	return "", false
}
func sortArrayNodes[T any](nodes []*pathTrieNode[T]) (res []*pathTrieNode[T]) {
	sort.Sort(priorities[T](nodes))
	return nodes
}

func (t *pathTrieNode[T]) backtraceWithTail(node *pathTrieNode[T], subs []string, params *[]string, index int, matchedPath *string, values *[]T) bool {
	if index == len(subs) {
		*matchedPath = node.fullPath
		*values = node.values
		return true
	}
	nodes := node.arrayChilds
	for _, subNode := range nodes {
		if subNode.subPath == subs[index] {
			if t.backtraceWithTail(subNode, subs, params, index+1, matchedPath, values) {
				return true
			}
		}

		if subNode.paramFlag {
			*params = append(*params, subs[index])
			if t.backtraceWithTail(subNode, subs, params, index+1, matchedPath, values) {
				return true
			}
			*params = (*params)[:len(*params)-1]
		}

		if subNode.wildFlag {
			*matchedPath = subNode.fullPath
			*values = subNode.values
			return true
		}
	}
	return false
}

func (t *pathTrieNode[T]) matchAll(subs []string) (results []MatchedResult[T]) {
	params := []string{}
	t.backtraceAll(t, subs, &params, 0, &results)
	return
}

func (t *pathTrieNode[T]) backtraceAll(node *pathTrieNode[T], subs []string, params *[]string, index int, results *[]MatchedResult[T]) {
	if index == len(subs) {
		result := MatchedResult[T]{
			Path:   node.fullPath,
			Params: *params,
			Values: node.values,
		}
		*results = append(*results, result)
		*params = []string{}
		return
	}
	// nodes := sortNodes(node.child)
	nodes := node.arrayChilds
	for _, subNode := range nodes {
		if subNode.subPath == subs[index] {
			t.backtraceAll(subNode, subs, params, index+1, results)
		} else if subNode.paramFlag {
			*params = append(*params, subs[index])
			t.backtraceAll(subNode, subs, params, index+1, results)
		} else if subNode.wildFlag {
			result := MatchedResult[T]{
				Path:   subNode.fullPath,
				Params: *params,
				Values: subNode.values,
			}
			*results = append(*results, result)
			*params = []string{}
		} else {
			continue
		}
	}
}
