# multiscanner

The program accepts a file with a list of IP addresses, divides them into packs.

For each bundle of addresses, an amazon-ec2 instance is raised from a pre-saved image

The nmap command is called on the instance, the result of the work is displayed in STDOUT


#### Config:

```
cp config-example.yaml config.yaml
```
