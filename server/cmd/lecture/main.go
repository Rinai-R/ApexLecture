package main

import (
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture/lectureservice"
	"log"
)

func main() {
	svr := lecture.NewServer(new(LectureServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
