package main

import (
	"bufio"
	"fmt"
	"net/http"
)

func main() {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDgwMTU0OTUsInN1YiI6MTkyNTQyMDA3MDA3OTU2OTkyMH0.xJxrqYNVZzKpM6oy5CzDa98ZGA1qE_mcrnENkFkIzxFDZIlCfigBG21Rs4q2AoQQS_LR4cF4ALthgLA5wytSBlm4vEH0a7IDMY3bHxkdCK3atbycn-cu_22VXgeXLFjfoI4U-KLdcsLzq-bObTjhv_nXDus2ekA10F53uRBHXqGKK0gRejLoCVyC5LDEoZaf-iXru9zcWv27SLL816Sm5DJDYBSdVEj4AxfCxR2rt8eQash5WxLSKvLffYmAcmOs-gY1S5Vha323vh0X6Itwu2RynVJUXRgOGiVzxgUIaINrFAOsM81YLQ01vdahxRIzgx9jOH9W2nNmJZQcRjuTQw"
	req, _ := http.NewRequest("GET", "http://localhost:10000/receive/1925898343783870464", nil)
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
