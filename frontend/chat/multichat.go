package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	for i := 0; i < 10000; i++ {
		token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDk0ODYzMjksInN1YiI6MTkzMjA4MTYwNTU1NjMxNDExMn0.CXdgxC6MquMEtj7LBl-SA1jhVPpQ3Ygdk3XMNGhwdgESkQac0bellY7UH14AvsTZGw5LztXw2cMxRX9MEQMSy6bGFGxOzs78w9qisPVYPEyNBhK-ChuVEkxe28ev81vE0wl7efbZ2CfWukvw-IrQ9Jgg51b51_8sEEzMis0BkU6QQyGxdslgF6nXWG2W2uW9XmWJxx-WpYyFGPVz1vkkMGZDbXfESTdU3grMfu3tfUA9fwHXoiyKc3LdjK36MZFBPdRosjhXUl2XIH37WQKX4IgNeqVnLVG_wYMLn9Ne_ePd_TcA99Ey_SP3jC2KU4Vwb5YU-mtmEAxsyKw-7egkuQ"
		data := `{
		"type":1,
		"message":"ciallo"
	}`
		req, _ := http.NewRequest("POST", "http://localhost:10000/chat/1932082377442471936", bytes.NewBuffer([]byte(data)))
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
