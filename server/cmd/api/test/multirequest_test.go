package test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestMultiGet(t *testing.T) {
	var check int64 = 0
	client := resty.New()
	GetNum := 20000
	wg := sync.WaitGroup{}
	wg.Add(GetNum)
	for i := 0; i < GetNum; i++ {
		go func() {
			res, _ := client.R().SetHeaders(map[string]string{}).Get("http://localhost:10000/ping")
			if res != nil && res.StatusCode() != 200 {
				t.Log("get status code", res.StatusCode())
				atomic.AddInt64(&check, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if check < 100 {
		t.Error("Fail to check TooManyRequests ", check)
	} else {
		t.Log("MultiGet test passed", check)
	}

}
