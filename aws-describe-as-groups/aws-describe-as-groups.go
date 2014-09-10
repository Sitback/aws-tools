/*
Small app to display any auto scaling group names in the region for this account

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/goamz/goamz/autoscaling"
	"github.com/goamz/goamz/aws"
	"github.com/gombadi/go-ini"
)

func main() {

	// pointers to objects we use to talk to AWS
	var as *autoscaling.AutoScaling
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
		regionName, ok = iniFile.Get("default", "region")
		if !ok {
			fmt.Printf("Error - unable to find AWS Region information\n")
			os.Exit(1)
		}
	}

	//  read secret key from command line or ini file
	// if either secret key or access key are not provided then read all info from ini file
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
	case "sa-east-1":
		as = autoscaling.New(auth, aws.SAEast)
	case "us-east-1":
		as = autoscaling.New(auth, aws.USEast)
	case "ap-northeast-1":
		as = autoscaling.New(auth, aws.APNortheast)
	case "us-west-2":
		as = autoscaling.New(auth, aws.USWest2)
	case "us-west-1":
		as = autoscaling.New(auth, aws.USWest)
	case "ap-southeast-1":
		as = autoscaling.New(auth, aws.APSoutheast)
	case "ap-southeast-2":
		as = autoscaling.New(auth, aws.APSoutheast2)
	default:
		fmt.Printf("Error - Sorry I can not find the url endpoint for region %s\n", regionName)
		os.Exit(1)
	}

	asGroups, err := as.DescribeAutoScalingGroups(nil,0,"")
	if err != nil {
		fmt.Println("\nBother, things are not looking good getting the Auto Scale Group Names...\n")
		log.Fatal(err)
	}

	fmt.Printf("Autoscaling groups in region %s\n", regionName)

	// extract the group names
	for asGroup := range asGroups.AutoScalingGroups {
		fmt.Printf("%s\n", asGroups.AutoScalingGroups[asGroup].AutoScalingGroupName)
	}

}
