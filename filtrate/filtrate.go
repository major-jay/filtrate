package filtrate

import (
	"bufio"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
)

var (
	Root *tree
)

type tree struct {
	next struct {
		sync.RWMutex
		m map[string]*tree
	} //子节点指针 {a:*tree}
	val    string //当前节点的字符，nil表示根节点
	back   *tree  //跳跃指针，也称为失败指针
	parent *tree  //父节点指针
	accept bool   //是否形成了一个完整的词汇，中间节点也可能为true
}

func InitTree(path string) *tree {
	var (
		words []string
		root  *tree
	)
	root = &tree{next: struct {
		sync.RWMutex
		m map[string]*tree
	}{m: make(map[string]*tree)}, val: "", back: nil, parent: nil, accept: false}
	f, err := os.Open(path)
	if err != nil {
		panic("读取错误")
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		word := strings.TrimSpace(line)
		isContain, _ := contain(word, words)
		if len(word) > 0 && !isContain {
			words = append(words, word)
			go addWord(root, word)
		}
	}
	fallbackAll(root)
	Root = root
	return root
}

//匹配所有分支的back
func fallbackAll(root *tree) {
	curExpands := mapToSlice(root.next.m)
	for len(curExpands) > 0 {
		var nextExpands []*tree
		for _, t := range curExpands {
			for _, c := range t.next.m {
				nextExpands = append(nextExpands, c)
			}
			go fallback(t)
		}
		curExpands = nextExpands
	}

}

func fallback(t *tree) {
	//找到父节点的back，从back的next里找是否有相同元素
	parent := t.parent
	back := parent.back
	for back != nil {
		child := back.next.m[t.val]
		if child != nil {
			t.back = child
			break
		}
		back = back.back
	}
}

//把类似slice的map转为slice
func mapToSlice(input map[string]*tree) []*tree {
	output := []*tree{}
	for _, value := range input {
		output = append(output, value)
	}
	return output
}

//向tree里面加入词汇
func addWord(root *tree, word string) {
	current := root
	root.next.Lock()
	for _, c := range word {
		b := string(c)
		isContain, _ := contain(b, current.next.m)
		if !isContain {
			current.next.m[b] = &tree{
				next: struct {
					sync.RWMutex
					m map[string]*tree
				}{m: make(map[string]*tree)},
				val:    b,
				back:   root,
				parent: current,
				accept: false,
			}
		}
		current = current.next.m[b]
	}
	current.accept = true
	root.next.Unlock()
}

// 判断obj是否在target中，target支持的类型arrary,slice,map
func contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

func Search(content string) []string {
	var offWords []string
	current := Root
	for _, c := range content {
		b := string(c)
		next := current.next.m[b]
		if next == nil {
			//处理失败->失败指针
			back := current.back
			for back != nil {
				next = back.next.m[b]
				if next != nil {
					break
				}
				back = back.back
			}
		} else {
			back := next
			for {
				if back.accept {
					offWords = append(offWords, compose(*back))
				}
				back = back.back
				if back == Root {
					break
				}
			}
			current = next
			continue
		}
		if Root.next.m[b] != nil {
			current = Root.next.m[b]
		} else {
			current = Root
		}

	}
	return offWords
}

func compose(node tree) string {
	word := ""
	for node.val != "" {
		word = node.val + word
		node = *node.parent
	}
	return word
}
