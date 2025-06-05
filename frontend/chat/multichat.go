package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	for i := 0; i < 10000; i++ {
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDkxMDM0NjUsInN1YiI6MTkzMDQ3NTcyNDA4NzUwMDgwMH0.fKm8oQHyGKtBqE8skdSlnOKkrorCN72pSUYiPYREUKCRK7cqGGO_-zIzIEHnuzi4slSAFUizyLxDh0rF3XwNGPUuAA9F4tpk32EvwXUmYqgVfk4vnehUiwddc3c3_VPXBUdY-5OV1x6t-9QcaBk_Wte9T2DyFoflfaDCTqTGdn6UomQHQbr1CEYaCaLu3tQBcjnSpYKIZ81rL0UumhA5rUcQHxjjUqcNRP2jDfCIxZ9KNQux1uq039kzxNPhXgzQBooWfHMl0MDTNRk0Vmnn16N3vlqDQRiJHMp-v655P2opnwYU-9FENXXbhh4AjpDCcYySNDNGovXLnmCQn6IOyQ"
		data := `{
		"type":1,
		"message":"ciallo"
	}`
		req, _ := http.NewRequest("POST", "http://localhost:10000/chat/1930477035466010624", bytes.NewBuffer([]byte(data)))
		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Println("Received: ", scanner.Text())
		}
	}
}
