package local

import (
	"bytes"
	"fmt"
	"log"
	"multiparser/provider"
	"os/exec"
	"sync"
	"time"
)

func NewProvider(buf *bytes.Buffer) *Provider {
	return &Provider{
		buf: buf,
		waitInstanceRequest: make(chan (chan<- provider.Instance), 1),
	}
}

type Provider struct {
	buf *bytes.Buffer
	waitInstanceRequest chan (chan<- provider.Instance)
}

func (p *Provider) GetInstance() <-chan provider.Instance {
	resp := make(chan provider.Instance)
	p.waitInstanceRequest <- resp
	return resp
}

func (p *Provider) Run() {
	for req := range p.waitInstanceRequest {
		go func(r chan<- provider.Instance) {
			log.Printf("Run: new instance request : do some work 5 sec...")
			time.Sleep(5 * time.Second)
			r <- NewInstance(p.buf)
		}(req)
	}
}

func NewInstance(buf *bytes.Buffer) *Instance {
	return &Instance{
		buf:  buf,
		Name: "Some name",
	}
}

type Instance struct {
	buf  *bytes.Buffer
	Name string
}

func (i *Instance) Execute(wg *sync.WaitGroup, cmdstr string) ([]byte, error) {
	defer wg.Done()
	b := &bytes.Buffer{}
	cmd := exec.Command(cmdstr)
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Execute: %s ", err)
	}
	return b.Bytes(), nil
}

func (i *Instance) String() string {
	return i.Name
}
