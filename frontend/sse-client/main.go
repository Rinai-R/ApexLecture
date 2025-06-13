package main

import (
	"bufio"
	"fmt"
	"net/http"
)

func main() {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDk0ODYzMjksInN1YiI6MTkzMjA4MTYwNTU1NjMxNDExMn0.CXdgxC6MquMEtj7LBl-SA1jhVPpQ3Ygdk3XMNGhwdgESkQac0bellY7UH14AvsTZGw5LztXw2cMxRX9MEQMSy6bGFGxOzs78w9qisPVYPEyNBhK-ChuVEkxe28ev81vE0wl7efbZ2CfWukvw-IrQ9Jgg51b51_8sEEzMis0BkU6QQyGxdslgF6nXWG2W2uW9XmWJxx-WpYyFGPVz1vkkMGZDbXfESTdU3grMfu3tfUA9fwHXoiyKc3LdjK36MZFBPdRosjhXUl2XIH37WQKX4IgNeqVnLVG_wYMLn9Ne_ePd_TcA99Ey_SP3jC2KU4Vwb5YU-mtmEAxsyKw-7egkuQ"
	req, _ := http.NewRequest("GET", "http://localhost:10000/receive/1932082377442471936", nil)
	req.Header.Set("Authorization", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println("收到：", scanner.Text())
	}
}
