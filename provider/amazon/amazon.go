package amazon

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"multiparser/provider"
	amazonConfig "multiparser/provider/amazon/config"
	"net"
	"strings"
	"sync"
	"time"
)

type Provider struct {
	config    amazonConfig.Amazon
	sshconfig *ssh.ClientConfig
	sess      *session.Session
	buf       *bytes.Buffer
	// EC2 service client
	svc                 *ec2.EC2
	waitInstanceRequest chan (chan<- provider.Instance)
}

func NewProvider(config amazonConfig.Amazon, keypair string, buf *bytes.Buffer) *Provider {
	pkdata, err := ioutil.ReadFile(keypair)
	if err != nil {
		log.Fatal(err)
	}
	pkey, err := ssh.ParsePrivateKey(pkdata)
	if err != nil {
		log.Fatal(err)
	}

	sshconf := &ssh.ClientConfig{
		User:            "ec2-user",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pkey),
		},
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.IdKey, config.SecretKey, ""),
	})
	if err != nil {
		log.Fatalf("NewProvider: Create session: %v", err)
	}
	return &Provider{
		buf:                 buf,
		sess:                sess,
		sshconfig:           sshconf,
		config:              config,
		svc:                 ec2.New(sess),
		waitInstanceRequest: make(chan (chan<- provider.Instance), 1),
	}
}

func (p *Provider) statusCheck(instance *ec2.Instance) bool {
	statusResp, err := p.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	if err != nil {
		log.Printf("statusCheck: error %s", err)
		return false
	}
	if len(statusResp.InstanceStatuses) == 0 {
		log.Printf("statusCheck: empty response")
		return false
	}
	systemStatus := *statusResp.InstanceStatuses[0].SystemStatus.Status

	log.Printf("statusCheck: instance %s status %s ", *instance.InstanceId, systemStatus)

	return strings.ToLower(systemStatus) == "ok"
}

// TODO: delete or cache running instances
func (p *Provider) runEC2Instance() *ec2.Instance {
	config := p.config.Instance
	svc := p.svc

	// Run Instance
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(config.ImageID),
		InstanceType:     aws.String(config.Type),
		MinCount:         aws.Int64(1),
		MaxCount:         aws.Int64(1),
		SecurityGroupIds: aws.StringSlice([]string{config.SecurityGroups.Id}),
	})
	if err != nil {
		fmt.Println("runEC2Instance: Could not create instance", err)
		return nil
	}
	instance := runResult.Instances[0]
	id := *instance.InstanceId
	fmt.Println("runEC2Instance: Created instance", id)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(config.PrefixName + "_" + id),
			},
		},
	})
	if errtag != nil {
		log.Println("runEC2Instance: Could not create tags for instance", id, errtag)
		return nil
	}

	fmt.Println("runEC2Instance: Successfully tagged instance")
	return instance
}

func (p *Provider) GetInstance() <-chan provider.Instance {
	ch := make(chan provider.Instance)
	p.waitInstanceRequest <- ch
	return ch
}

func (p *Provider) Run() {
	for req := range p.waitInstanceRequest {
		go func(r chan<- provider.Instance) {
			ec2instance := p.runEC2Instance()
			// Wait status check OK return
			for {
				log.Printf("Run: waiting instance starting %s ", *ec2instance.InstanceId)
				log.Printf("Run: Wait 5 seconds...")
				time.Sleep(5 * time.Second)
				// TODO: timeout
				if p.statusCheck(ec2instance) {
					out, _ := p.svc.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: []*string{ec2instance.InstanceId}})
					// TODO: not safe
					res := out.Reservations[0].Instances[0]
					r <- &Instance{
						buf:       p.buf,
						PublicIP:  *res.PublicIpAddress,
						sshconfig: p.sshconfig,
						PublicDNS: *res.PublicDnsName,
						I:         ec2instance,
					}
					return
				}
				continue
			}
		}(req)
	}
}

type Instance struct {
	buf                 *bytes.Buffer
	sshconfig *ssh.ClientConfig
	PublicIP, PublicDNS string
	I                   *ec2.Instance
}

func (i Instance) String() string {
	return i.I.String()
}

func (i *Instance) Execute(wg *sync.WaitGroup, cmd string) ([]byte, error) {
	defer wg.Done()

	b := &bytes.Buffer{}

	client, err := ssh.Dial("tcp", net.JoinHostPort(i.PublicIP, "22"), i.sshconfig)
	if err != nil {
		return nil, fmt.Errorf("Execute: %s ", err)
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Execute: %s ", err)
	}
	defer sess.Close()

	sess.Stdout = b
	log.Printf("Execute: instance %s; run cmd: %s ", *i.I.InstanceId, cmd)

	if err = sess.Run(cmd); err != nil {
		return nil, fmt.Errorf("Execute: %s ", err)
	}
	return b.Bytes(), nil
}
