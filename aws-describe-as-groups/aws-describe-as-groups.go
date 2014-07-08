/*
Small app to display any auto scaling group names in the region for this account

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/autoscaling"
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
	var as *autoscaling.AutoScaling
	// storage for commandline args
	var regionName, awsKey, awsSecret, awsRegion string

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
		fmt.Printf("Error - Sorry I can not find the url endpoint for region %s\n", awsRegion)
		os.Exit(1)
	}

	asGroups, err := as.DescribeAutoScalingGroups(nil)
	if err != nil {
		fmt.Println("\nBother, things are not looking good getting the Auto Scale Group Names...\n")
		log.Fatal(err)
	}

	fmt.Printf("Autoscaling groups in region %s\n", awsRegion)

	// extract the group names
	for asGroup := range asGroups.AutoScalingGroups {
		fmt.Printf("%s\n", asGroups.AutoScalingGroups[asGroup].AutoScalingGroupName)
	}

}
