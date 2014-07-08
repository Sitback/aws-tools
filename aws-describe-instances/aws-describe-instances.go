/*
Small app to display basic instance info

*/

package main

import (
	"flag"
	"fmt"
	"os"
	"log"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	"github.com/gombadi/go-ini"
)

// cmdline flag if we want verbose output
var verbose bool

// loadAWSCredentials function will first try and read from the env variables
// and if nothing found then try from the standard file then from the command line
func loadAWSCredentials(awsKey, awsSecret, regionName string) {

	var awsSecretf, awsKeyf, regionNamef string
	var ok bool

	if len(os.Getenv("AWS_SECRET_ACCESS_KEY")) > 5 && len(os.Getenv("AWS_ACCESS_KEY_ID")) > 5 {
		// all must be good so lets try these credentials
		if verbose {
			fmt.Printf("Info - Using AWS credentials from the environment\n")
		}
		return
	}

	// read the standard AWS ini file in case it is needed
	iniFile, err := ini.LoadFile(os.Getenv("HOME") + "/.aws/credentials")

	if err != nil {
		// if we get an error with the standard name try the previous name
		iniFile, err = ini.LoadFile(os.Getenv("HOME") + "/.aws/config")
	}

	awsSecretf, ok = iniFile.Get("default", "aws_secret_access_key")
	if !ok {
		// failed to read from file so last chance is command line
		if awsSecret == "xxxx" {
			log.Fatalf("Error - unable to find AWS Secret Key information\n")
		} else {
			os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecret)
		}

	} else {
		os.Setenv("AWS_SECRET_ACCESS_KEY", awsSecretf)
	}

	awsKeyf, ok = iniFile.Get("default", "aws_access_key_id")
	if !ok {
		// failed to read from file so last chance is command line
		if awsKey == "xxxx" {
			log.Fatalf("Error - unable to find AWS Access Key information\n")
		} else {
			os.Setenv("AWS_ACCESS_KEY_ID", awsKey)
		}

	} else {
		os.Setenv("AWS_ACCESS_KEY_ID", awsKeyf)
	}

	// region is not a standard for the account so if not provided on
	// command line then pull from the config file
	if regionName == "xxxx" {
		regionNamef, ok = iniFile.Get("default", "region")
		if !ok {
			log.Fatalf("Error - unable to find AWS Region information\n")
		} else {
			os.Setenv("AWS_ACCESS_REGION", regionNamef)
		}
	} else {
		os.Setenv("AWS_ACCESS_REGION", regionName)
	}
	if verbose {
		fmt.Printf("Info - Using AWS credentials from config file and connecting to region %s\n", os.Getenv("AWS_ACCESS_REGION"))
	}

}




func main() {

	// pointers to objects we use to talk to AWS
	var e *ec2.EC2
	// storage for commandline args
	var regionName, awsKey, awsSecret, awsRegion  string

	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key")
	flag.Parse()


	// load the AWS credentials from the environment or from the standard file
	loadAWSCredentials(awsKey, awsSecret, regionName)

	// pull the region from the environment
	awsRegion = os.Getenv("AWS_ACCESS_REGION")

	// Pull the access details from environment variables
	// AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	auth, err := aws.EnvAuth()
	if err != nil {
		fmt.Printf("Error unable to find Access and/or Secret key\n")
		os.Exit(1)
	}

	// create the objects we will use to talk to AWS
	switch awsRegion {
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
		fmt.Printf("Error - Sorry I can not find the url endpoint for region %s\n", awsRegion)
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
