package main

import (
	quiz "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz/quizservice"
	"log"
)

func main() {
	svr := quiz.NewServer(new(QuizServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
