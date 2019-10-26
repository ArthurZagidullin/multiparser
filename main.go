package main

import (
	"fmt"
	"log"
	"multiparser/config"
	"multiparser/provider/amazon"
	"os/exec"
)

//1. Принять список
//2. Разбить на пачки
//3. Подготовить инстансы выполнения
//4. Отправлять пачки в инстансы
//   на обработку по мере их готовности
func main() {
	cfg := config.Config{}
	if err := cfg.Load("./config.yaml"); err != nil {
		panic(err)
	}

	//pvd := local.NewProvider()
	//go pvd.Run()
	//
	//inst := <-pvd.GetInstance()
	//log.Printf("%s", inst)
	//
	//res, err := inst.Execute(exec.Command("ls", "-l", "-1"))
	//if err != nil {
	//	log.Fatalf("Execute command in Instance: %s", err)
	//}
	//log.Printf("Execute result: %s ", res)

	amazoncfg := cfg.Providers.Amazon

	pvd := amazon.NewProvider(amazoncfg)
	go pvd.Run()

	inst := <-pvd.GetInstance()

	fmt.Printf("\n%s\n", inst)

	cmd := exec.Command("ssh", "-o \"StrictHostKeyChecking=no\"","-i" + amazoncfg.Instance.SecurityGroups.KeyPair)

	res, err := inst.Execute(cmd)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("\nResutlt: %s \n", res)

	//inst.Execute(exec.Command())
	//
	//inst := pvd.GetInstance()
	//test()
	fmt.Scanln()
}
