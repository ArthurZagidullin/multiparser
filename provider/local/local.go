package local

import (
	"bytes"
	"fmt"
	"log"
	"multiparser/provider"
	"os/exec"
	"time"
)

func NewProvider() *Provider {
	return &Provider{
		waitInstanceRequest: make(chan (chan<- provider.Instance), 1),
	}
}

type Provider struct {
	waitInstanceRequest chan (chan<- provider.Instance)
}

func (p *Provider ) GetInstance() <-chan provider.Instance {
	resp := make(chan provider.Instance)
	p.waitInstanceRequest <- resp
	return resp
}

func (p *Provider) Run() {
	for req := range p.waitInstanceRequest {
		go func(r chan<- provider.Instance) {
			log.Printf("Run: new instance request : do some work 5 sec...")
			time.Sleep(5 * time.Second)
			r <- NewInstance()
		}(req)
	}
}

func NewInstance() *Instance {
	return &Instance{
		Buf: &bytes.Buffer{},
		Name: "Some name",
	}
}

type Instance struct {
	Buf *bytes.Buffer
	Name string
}

func (i *Instance) Execute (cmd *exec.Cmd) ([]byte, error) {
	cmd.Stdout = i.Buf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Execute: %s ", err)
	}
	return i.Buf.Bytes(), nil
}

func (i *Instance) String() string {
	return i.Name
}
