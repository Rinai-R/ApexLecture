package setequal

import "github.com/cloudwego/kitex/pkg/klog"

// 用于比较选择题的答案
func Compare(a, b []int8) bool {
	mp1 := make(map[int8]bool)
	mp2 := make(map[int8]bool)
	for _, v := range a {
		mp1[v] = true
	}
	for _, v := range b {
		mp2[v] = true
	}
	klog.Info("mp1: ", mp1, " mp2: ", mp2)
	for k, va := range mp1 {
		if vb, ok := mp2[k]; !ok || vb != va {
			return false
		}
	}
	return true
}
