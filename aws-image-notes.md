Idea: Boot a preconfigured AMI with consul installed, that reads the executable produces by codeship.


		
1. Create a policy

ds-codeship readers:

		{
		  "Version": "2012-10-17",
		  "Statement": [
		    {
		      "Effect": "Allow",
		      "Action": [
		        "s3:Get*",
		        "s3:List*"
		      ],
		      "Resource": [ "arn:aws:s3:::ds-codeship/*"]
		    }
		  ]
		}
		
Create a role:

xavi-deploy
AWS Service Role > Amazon EC2

Role ARN: arn:aws:iam::930295567417:role/xavi-deploy

Create customer image:

Baseline is Ubuntu Server 14.04 LTS (HVM), SSD Volume Type - ami-9a562df2

sudo apt-get -yqq update
sudo apt-get install -y zip
sudo apt-get install -y wget
cd /usr/local/bin
sudo wget https://dl.bintray.com/mitchellh/consul/0.4.1_linux_amd64.zip
sudo unzip 0.4.1_linux_amd64.zip
sudo rm *.zip

sudo mkdir -p /opt/consul/ui
cd /opt/consul/ui
sudo wget wget https://dl.bintray.com/mitchellh/consul/0.4.1_web_ui.zip
sudo unzip 0.4.1_web_ui.zip

sudo mkdir -p /var/consul
sudo mkdir -p /etc/consul.d/server

sudo vi /etc/consul.d/server/config.json and add the following content:

{
    "bootstrap_expect": 1,
    "server": true,
    "datacenter": "aws1",
    "data_dir": "/var/consul",
    "ui_dir": "/opt/consul/ui/dist",
    "log_level": "INFO",
    "enable_syslog": true
}

cd /etc/init

sudo vi consul.conf and add the following content:

description "Consul server process"

start on (local-filesystems and net-device-up IFACE=eth0)
stop on runlevel [!12345]

respawn


exec consul agent -config-dir /etc/consul.d/server
sudo python get-pip.py
sudo pip install awscli


Command line tools
curl "https://bootstrap.pypa.io/get-pip.py" -o "get-pip.py"


AMI: ubuntu-consul-aws-baseline - ami-ac5e08c4

Launch script

#!/bin/sh
aws s3 cp s3://ds-codeship/xavi /usr/local/bin
chmod +x /usr/local/bin/xavi


Security Group
TCP ports inside subnet 10.0.0.0/24 - 8300, 8301, 8302, 8400, 8500, 8600
UDP 8301, 8302, 8600
See Ports Used Section - http://www.consul.io/docs/agent/options.html

Once the image is build it can be launched from the command line like:

		aws --region us-east-1 ec2 run-instances \
		--profile codeshipper --image-id ami-ac5e08c4 --key-name FidoKeyPair \
		--security-group-ids sg-118f3d75 \
		--user-data IyEvYmluL3NoDQphd3MgczMgY3AgczM6Ly9kcy1jb2Rlc2hpcC94YXZpIC91c3IvbG9jYWwvYmluDQpjaG1vZCAreCAvdXNyL2xvY2FsL2Jpbi94YXZp \
		--instance-type t2.micro --placement AvailabilityZone=us-east-1c \
		--subnet-id  subnet-fcbd7fd7 --iam-instance-profile Arn=arn:aws:iam::930295567417:instance-profile/xavi-deploy \
		--count 1 --associate-public-ip-address

Note you need to set HTTP_PROXY and HTTPS_PROXY. Also note the IAM policy to run the above is:

{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "arn:aws:s3:::ds-codeship/*"
    }
  ]
}

{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt1424299617000",
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "iam:AddRoleToInstanceProfile",        
        "iam:CreateInstanceProfile",
        "iam:CreateRole",
        "iam:DeleteInstanceProfile",
        "iam:DeleteRole",
        "iam:DeleteRolePolicy",
        "iam:GetRole",
        "iam:PassRole",
        "iam:PutRolePolicy",
        "iam:RemoveRoleFromInstanceProfile"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}

	


This might not be a minimal entitlements.

In code ship, set the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables,
and configure the following deployment command:

pip install awscli
aws --region us-east-1 ec2 run-instances --image-id ami-ac5e08c4 --key-name FidoKeyPair --security-group-ids sg-118f3d75 --user-data IyEvYmluL3NoDQphd3MgczMgY3AgczM6Ly9kcy1jb2Rlc2hpcC94YXZpIC91c3IvbG9jYWwvYmluDQpjaG1vZCAreCAvdXNyL2xvY2FsL2Jpbi94YXZp --instance-type t2.micro --placement AvailabilityZone=us-east-1c --subnet-id  subnet-fcbd7fd7 --iam-instance-profile Arn=arn:aws:iam::930295567417:instance-profile/xavi-deploy --count 1 --associate-public-ip-address

