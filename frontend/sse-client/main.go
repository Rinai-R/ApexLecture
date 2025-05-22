package main

import (
	"bufio"
	"fmt"
	"net/http"
)

func main() {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDc4OTgwOTUsInN1YiI6MTkyNTQyMDA3MDA3OTU2OTkyMH0.manKPlNCqyduev5ogOW21618Mkr5aDEpleOI08fAPjEBLZtQFPDpqmxORgh9WLW84zlF2TTVIoLwE4h3F7LW-0fmg0Qa5z3MSuew7iZjPsUI55MBqxPuG9dmCyj7sqGD-4Fj4vxoNN2C8pojmpWoVxLIUEnKbpUXOjYQiB6ss77VpKKp7FJcRRe2YPnabyJ5PXJjnSIW5Oi153xz997yXAkFlZDo1sr1cItNey1KXJmPdWgXDxsJDVXjA4JUm7H1bi5g5_8qmeOlI1vUAAiaozZL-7wh-fWxCFyl1n-6pqvOHa98MDpljfBSK0hZGNo2fBYgldSbDlZa35nJKQxktA"
	req, _ := http.NewRequest("GET", "http://localhost:10000/receive/1925420168754769920", nil)
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
