package provider

import (
	"fmt"
	"os/exec"
)

// Instance среда выполнение ssh команд
type Instance interface {
	fmt.Stringer
	Execute(cmd *exec.Cmd) ([]byte, error)
}

// Provider задача подготовить instance
type Provider interface {
	GetInstance() <- chan  Instance
	Run()
}
