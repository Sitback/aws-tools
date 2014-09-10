/*
This application will display the private ip addresses for each instance in
an Amazon Auto Scaling Group

	asg-servers -r <regionname> -a <autoscaling group name>

	Command line arguments -
	-r AWS region name to use
	-a AWS Auto Scaling Group Name. Read from environment if not supplied
	-k AWS Access Key. Read from environment if not supplied
	-s AWS Secret key Read from environment if not supplied

*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goamz/goamz/autoscaling"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
	"github.com/gombadi/go-ini"
)

func main() {

	// pointers to objects we use to talk to AWS
	var e *ec2.EC2
	var as *autoscaling.AutoScaling

	// storage for commandline args
	var regionName, awsKey, awsSecret, asgName string
	var ok bool

	flag.StringVar(&asgName, "a", "xxxx", "AWS Auto scale group name")
	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key")
	flag.Parse()

	// read the standard AWS ini file in case it is needed
	iniFile, err := ini.LoadFile(os.Getenv("HOME") + "/.aws/config")

	if regionName == "xxxx" {
		regionName, ok = iniFile.Get("default", "region")
		if !ok {
			fmt.Printf("Error - unable to find AWS Region information\n")
			os.Exit(1)
		}
	}

	// if any value is not supplied on the command line then try and read the values
	// from the ini file
	if awsSecret == "xxxx" || awsKey == "xxxx" {
		awsSecret, ok = iniFile.Get("default", "aws_secret_access_key")
		if !ok {
			fmt.Printf("Error - unable to find AWS Secret Key information\n")
			os.Exit(1)
		}

		awsKey, ok = iniFile.Get("default", "aws_access_key_id")
		if !ok {
			fmt.Printf("Error - unable to find AWS Access Key information\n")
			os.Exit(1)
		}
	}
	// store auth info in environment
	os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecret)
	os.Setenv("AWS_ACCESS_KEY_ID", awsKey)

	if asgName == "xxxx" {
		fmt.Printf("Error - No AWS Auto scale name provided\n")
		os.Exit(1)
	}

	// Pull the access details from environment variables
	// AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	auth, err := aws.EnvAuth()
	if err != nil {
		fmt.Printf("Error unable to find Access and/or Secret key\n")
		os.Exit(1)
	}

	// create the objects we will use to talk to AWS
	switch regionName {
	case "eu-west-1":
		as = autoscaling.New(auth, aws.EUWest)
		e = ec2.New(auth, aws.EUWest)
	case "sa-east-1":
		as = autoscaling.New(auth, aws.SAEast)
		e = ec2.New(auth, aws.SAEast)
	case "us-east-1":
		as = autoscaling.New(auth, aws.USEast)
		e = ec2.New(auth, aws.USEast)
	case "ap-northeast-1":
		as = autoscaling.New(auth, aws.APNortheast)
		e = ec2.New(auth, aws.APNortheast)
	case "us-west-2":
		as = autoscaling.New(auth, aws.USWest2)
		e = ec2.New(auth, aws.USWest2)
	case "us-west-1":
		as = autoscaling.New(auth, aws.USWest)
		e = ec2.New(auth, aws.USWest)
	case "ap-southeast-1":
		as = autoscaling.New(auth, aws.APSoutheast)
		e = ec2.New(auth, aws.APSoutheast)
	case "ap-southeast-2":
		as = autoscaling.New(auth, aws.APSoutheast2)
		e = ec2.New(auth, aws.APSoutheast2)
	default:
		fmt.Printf("Error - Sorry I can not find the url endpoint for region %s\n", regionName)
		os.Exit(1)
	}

	//
	asgNames := []string{}
	if len(asgName) > 0 {
		asgNames = append(asgNames, asgName)
	}

	groupResp, err := as.DescribeAutoScalingGroups(asgNames,0,"")
	if err != nil {
		fmt.Println("\nBother, things are not looking good getting the autoscale details...\n")
		panic(err)
	}

	instanceSlice := []string{}

	if len(groupResp.AutoScalingGroups) < 1 {
		fmt.Printf("No Auto Scale Group info found for %s.\n", asgNames[0])
		os.Exit(1)
	}

	for asGroup := range groupResp.AutoScalingGroups {
		for instance := range groupResp.AutoScalingGroups[asGroup].Instances {
			// extract the instanceid's from the auto scale details and append to a slice
			instanceSlice = append(instanceSlice, groupResp.AutoScalingGroups[asGroup].Instances[instance].InstanceId)
		}
	}

	if len(instanceSlice) < 1 {
		fmt.Printf("No instances in auto scale group %s.\n", asgNames[0])
		os.Exit(1)
	}

	instanceResp, err := e.DescribeInstances(instanceSlice, nil)
	if err != nil {
		fmt.Println("\nBother, things are not looking good getting the Instance private ip addresses...\n")
		panic(err)
	}

	// extract the private ip address from the instance struct stored in the reservation
	for reservation := range instanceResp.Reservations {
		for instance := range instanceResp.Reservations[reservation].Instances {
			fmt.Printf("%s\n",
				instanceResp.Reservations[reservation].Instances[instance].PrivateIPAddress)
		}
	}

}
