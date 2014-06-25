/*
Small app to display basic instance info

*/

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	"github.com/sitback/go-ini"
)

func main() {

	// pointers to objects we use to talk to AWS
	var e *ec2.EC2
	// storage for commandline args
	var regionName, awsKey, awsSecret string
	var ok bool

	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key")
	flag.Parse()

	// read the standard AWS ini file in case it is needed
	iniFile, err := ini.LoadFile(os.Getenv("HOME") + "/.aws/config")

	// use any values not supplied on the command line
	if regionName == "xxxx" {
		regionName, ok = iniFile.Geti("default", "region")
		if !ok {
			fmt.Printf("Error - unable to find AWS Region information\n")
			os.Exit(1)
		}
	}

	//  read secret key from command line or ini file
	// if either secret key or access key are not provided then read all info from ini file
	if awsSecret == "xxxx" || awsKey == "xxxx" {
		awsSecret, ok = iniFile.Geti("default", "AWS_SECRET_ACCESS_KEY")
		if !ok {
			fmt.Printf("Error - unable to find AWS Secret Key information\n")
			os.Exit(1)
		}

		awsKey, ok = iniFile.Geti("default", "AWS_ACCESS_KEY_ID")
		if !ok {
			fmt.Printf("Error - unable to find AWS Access Key information\n")
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

	var instanceName string

	// extract the private ip address from the instance struct stored in the reservation
	for reservation := range instanceResp.Reservations {
		for instance := range instanceResp.Reservations[reservation].Instances {
			for tag := range instanceResp.Reservations[reservation].Instances[instance].Tags {
				if instanceResp.Reservations[reservation].Instances[instance].Tags[tag].Key == "Name" {
					instanceName = instanceResp.Reservations[reservation].Instances[instance].Tags[tag].Value
					break
				} else {
					instanceName = "Unknown"
				}
			}
			fmt.Printf("Instance: %s\tName: %s\tstate: %s\tType: %s\tAVzone: %s\tPublicIP: %s\tPrivateIP: %s\n",
				instanceResp.Reservations[reservation].Instances[instance].InstanceId,
				instanceName,
				instanceResp.Reservations[reservation].Instances[instance].State.Name,
				instanceResp.Reservations[reservation].Instances[instance].InstanceType,
				instanceResp.Reservations[reservation].Instances[instance].AvailabilityZone,
				instanceResp.Reservations[reservation].Instances[instance].IPAddress,
				instanceResp.Reservations[reservation].Instances[instance].PrivateIPAddress)
		}
	}

}
