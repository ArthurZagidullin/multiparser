package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"multiparser/config"
	"multiparser/provider/amazon"
	"os"
	"sync"
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

	storage := &bytes.Buffer{}
	pvd := amazon.NewProvider(cfg.Providers.Amazon, cfg.Providers.Amazon.Instance.SecurityGroups.KeyPair, storage)
	go pvd.Run()

	iplist, err := readLines(cfg.Common.Iplist)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	for _, ippack := range dividePack(iplist, cfg.Common.PackLimit) {
		fmt.Printf("IP Pack: %v\n", ippack)
		wg.Add(1)
		go func(wg *sync.WaitGroup, pack []string) {
			defer wg.Done()
			wgcmd := &sync.WaitGroup{}
			inst := <-pvd.GetInstance()
			for _, ip := range pack {
				wgcmd.Add(1)
				go func(wgcmd *sync.WaitGroup, ip string) {
					b, err := inst.Execute(wgcmd, "nmap -F "+ip)
					if  err != nil {
						log.Printf("Error: %s ", err)
						return
					}
					fmt.Printf("\nResult %s:\n %s", ip, b)
				}(wgcmd, ip)
			}
			wgcmd.Wait()
		}(wg, ippack)
	}

	wg.Wait()
	fmt.Printf("\n-------------\nStorage Result: %s \n", storage.String())
}

func dividePack(slice []string, size int) [][]string {
	var divided [][]string

	for i := 0; i < len(slice); i += size {
		end := i + size

		if end > len(slice) {
			end = len(slice)
		}

		divided = append(divided, slice[i:end])
	}

	return divided
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
