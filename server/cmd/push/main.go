package main

import (
	push "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push/pushservice"
	"log"
)

func main() {
	svr := push.NewServer(new(PushServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
