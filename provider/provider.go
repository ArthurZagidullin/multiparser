package provider

import (
	"fmt"
	"sync"
)

// Instance среда выполнение ssh команд
type Instance interface {
	fmt.Stringer
	Execute(wg *sync.WaitGroup, cmd string) ([]byte, error)
}

// Provider задача подготовить instance
type Provider interface {
	GetInstance() <-chan Instance
	Run()
}
