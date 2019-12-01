package helper

import "log"

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalln("error: ", err, "msg: ", msg)
	}

}
