package sensitive

import "github.com/importcjj/sensitive"

type FilterImpl struct {
	filter *sensitive.Filter
}

func NewFilter() *FilterImpl {
	filter := sensitive.New()
	// 加载自定义词库
	filter.LoadWordDict("./server/shared/sensitive/word.txt")
	return &FilterImpl{
		filter: filter,
	}
}

func (f *FilterImpl) ReplaceWithChar(text string, char rune) string {
	return f.filter.Replace(text, char)
}

func (f *FilterImpl) MultiReplaceWithChar(text []string, replace rune) []string {
	res := make([]string, 0)
	for _, t := range text {
		res = append(res, f.filter.Replace(t, replace))
	}
	return res
}
