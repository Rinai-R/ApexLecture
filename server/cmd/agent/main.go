package main

import (
	agent "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/agent/agent"
	"log"
)

func main() {
	svr := agent.NewServer(new(AgentImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
