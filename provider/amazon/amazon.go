package amazon

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"multiparser/provider"
	amazonConfig "multiparser/provider/amazon/config"
	"os/exec"
	"strings"
	"time"
)


type Provider struct {
	sess *session.Session
	// EC2 service client
	svc *ec2.EC2
	config amazonConfig.Amazon

	waitInstanceRequest chan (chan<- provider.Instance)
}

func NewProvider(config amazonConfig.Amazon) *Provider {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.IdKey, config.SecretKey, ""),
	})
	if err != nil {
		log.Fatalf("NewProvider: Create session: %v", err)
	}
	return &Provider{
		sess: sess,
		config: config,
		svc: ec2.New(sess),
		waitInstanceRequest: make(chan (chan<- provider.Instance), 1),
	}
}

func (p *Provider) statusCheck(instance *ec2.Instance) bool {
	statusResp, err := p.svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	if err != nil {
		log.Printf("statusCheck: error %s", err )
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
		ImageId:      aws.String(config.ImageID),
		InstanceType: aws.String(config.Type),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
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

func (p *Provider) GetInstance() <- chan provider.Instance {
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
					out, _ := p.svc.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds:[]*string{ec2instance.InstanceId}})
					// TODO: not safe
					res := out.Reservations[0].Instances[0]
					//out.
					r <- &Instance{
						PublicIP: *res.PublicIpAddress,
						PublicDNS: *res.PublicDnsName,
						I: ec2instance,
					}
					return
				}
				continue
			}
		}(req)
	}
}


type Instance struct {
	PublicIP, PublicDNS string
	I *ec2.Instance
}

func (i Instance) String() string {
	return i.I.String()
}

func (i *Instance) Execute(cmd *exec.Cmd) ([]byte, error) {
	//cmd.
	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	//exec.Command("ssh","-i" + amazoncfg.Instance.SecurityGroups.KeyPair,)
	sshpath := fmt.Sprintf("ec2-user@%s", i.PublicIP, )
	cmd.Args = append(cmd.Args, sshpath, "pwd")

	log.Printf("Execute: %s", strings.Join(cmd.Args, " "))

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Execute: %s ", err)
	}
	return buf.Bytes(), nil
}