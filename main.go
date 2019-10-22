package main

import (
	"log"
	"multiparser/config"
)

//1. Принять список
//2. Разбить на пачки
//3. Подготовить инстансы выполнения
//4. Отправлять пачки в инстансы
//   на обработку по мере их готовности
func main()  {
	cfg := config.Config{}
	if err := cfg.Load("./config.yaml"); err != nil {
		panic(err)
	}
	log.Printf("Config: %+v ", cfg)


}
