/*
This application will stop any instance on the account if it has a tag autostop
and if it is not running.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gombadi/go-ini"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

func main() {

	// pointers to objects we use to talk to AWS
	var e *ec2.EC2

	instanceSlice := []string{}

	// storage for commandline args
	var regionName, awsKey, awsSecret string
	var ok, quiet bool

	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key")

	flag.BoolVar(&quiet, "q", false, "Suppress no instances found message")
	flag.Parse()

	// read the standard AWS ini file in case it is needed
	iniFile, err := ini.LoadFile(os.Getenv("HOME") + "/.aws/credentials")

	// if any value is not provided on the command line then read all values
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
	if regionName == "xxxx" {
		regionName, ok = iniFile.Get("default", "region")
		if !ok {
			fmt.Printf("Error - unable to find AWS Region information\n")
			os.Exit(1)
		}
	}

	// store auth info in environment
	os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecret)
	os.Setenv("AWS_ACCESS_KEY_ID", awsKey)

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
		e = ec2.New(auth, aws.EUWest)
	case "sa-east-1":
		e = ec2.New(auth, aws.SAEast)
	case "us-east-1":
		e = ec2.New(auth, aws.USEast)
	case "ap-northeast-1":
		e = ec2.New(auth, aws.APNortheast)
	case "us-west-2":
		e = ec2.New(auth, aws.USWest2)
	case "us-west-1":
		e = ec2.New(auth, aws.USWest)
	case "ap-southeast-1":
		e = ec2.New(auth, aws.APSoutheast)
	case "ap-southeast-2":
		e = ec2.New(auth, aws.APSoutheast2)
	default:
		fmt.Printf("Error - Sorry I can not find the url endpoint for region %s\n", regionName)
		os.Exit(1)
	}

	instanceResp, err := e.DescribeInstances(nil, nil)
	if err != nil {
		fmt.Println("\nBother, things are not looking good getting the Instance private ip addresses...\n")
		panic(err)
	}

	// extract the instanceId with autostop tags and state running
	for reservation := range instanceResp.Reservations {
		for instance := range instanceResp.Reservations[reservation].Instances {
			for tag := range instanceResp.Reservations[reservation].Instances[instance].Tags {
				if instanceResp.Reservations[reservation].Instances[instance].Tags[tag].Key == "autostop" &&
					instanceResp.Reservations[reservation].Instances[instance].State.Name == "running" {
					// Found an instance that needs stopping
					instanceSlice = append(instanceSlice, instanceResp.Reservations[reservation].Instances[instance].InstanceId)
				}
			}
		}
	}

	// make sure we don't stop everything on the account
	if len(instanceSlice) < 1 {
		if !quiet {
			fmt.Printf("No autostop instances found\n")
		}
		os.Exit(0)
	}

	// oh I wish people would use consistant types in functions
	stopinstanceResp, err := e.StopInstances(strings.Join(instanceSlice, ", "))
	if err != nil {
		fmt.Println("\nBother, things are not looking good stopping the instances...\n")
		panic(err)
	}

	for statechange := range stopinstanceResp.StateChanges {
		fmt.Printf("InstanceId: %s\t\tPrevious state: %s\t\tNew State: %s\n",
			stopinstanceResp.StateChanges[statechange].InstanceId,
			stopinstanceResp.StateChanges[statechange].PreviousState.Name,
			stopinstanceResp.StateChanges[statechange].CurrentState.Name)
	}
}
