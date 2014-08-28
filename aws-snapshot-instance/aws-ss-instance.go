/*
This application will snapshot an amazon instance and also give it
tags of Key:Date, Value:todays date and key:Name, Value:same as existing
Name Tag or use instanceId

Command line options -
-r AWS region to talk to
-k AWS Access Key
-s AWS secret key
-i ami-id to be removed
-v verbose mode
-d <days> or
-a <days> autodelete mode enabled. Delete images older that this days.


Auto delete - When this mode is active the program will scan the current
account for any AMI that has an autodelete tag and that the are older than
days old then deregister the AMI and delete any associated snapshots.

*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
	"github.com/gombadi/go-ini"
	"github.com/gombadi/go-rate"
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
			log.Fatalf("error - unable to find AWS secret sey information\n")
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
			log.Fatalf("Error - unable to find AWS access key information\n")
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
			log.Fatalf("error - unable to find AWS region information\n")
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

func getBkupInstances(e *ec2.EC2, bkupId string) (bkupInstances []ec2.CreateImage) {

	var theInstance ec2.CreateImage

	instanceSlice := []string{}
	filter := ec2.NewFilter()

	// if instance id provided use it else search for tags autobkup
	if len(bkupId) > 0 {
		instanceSlice = append(instanceSlice, bkupId)
		filter = nil
	} else {
		filter.Add("tag-key", "autobkup")
		instanceSlice = nil
	}

	instanceResp, err := e.DescribeInstances(instanceSlice, filter)
	if err != nil {
		log.Fatalf("\nerror getting the instance details with tag autobkup\n%v\n", err)
	}

	// for any instance found extract tag name and instanceid
	for reservation := range instanceResp.Reservations {
		for instance := range instanceResp.Reservations[reservation].Instances {
			for tag := range instanceResp.Reservations[reservation].Instances[instance].Tags {
				if instanceResp.Reservations[reservation].Instances[instance].Tags[tag].Key == "Name" {
					// name of the created AMI must be unique so add the Unix Epoch
					theInstance.Name = instanceResp.Reservations[reservation].Instances[instance].Tags[tag].Value + "-" + strconv.FormatInt(time.Now().Unix(), 10)
					break
				} else {
					theInstance.Name = instanceResp.Reservations[reservation].Instances[instance].InstanceId + "-" + strconv.FormatInt(time.Now().Unix(), 10)
				}
			}
			theInstance.InstanceId = instanceResp.Reservations[reservation].Instances[instance].InstanceId
			theInstance.NoReboot = true
			// append details on this instance to the slice
			bkupInstances = append(bkupInstances, theInstance)
		}
	}
	return
}

func ssInstance(e *ec2.EC2, abkupInstance *ec2.CreateImage) {

	createImageResp, err := e.CreateImage(abkupInstance)
	if err != nil {
		log.Printf("non-fatal error creating the AMI image: %v\n", err)
		// any problems and out of here as createImageResp is invalid
		return
	}
	// store the creation time in the tag so it can be checked during auto cleanup
	tags := []ec2.Tag{ec2.Tag{Key: "autocleanup", Value: strconv.FormatInt(time.Now().Unix(), 10)}}
	_, err = e.CreateTags([]string{createImageResp.ImageId}, tags)
	if err != nil {
		log.Printf("non-fatal error adding autocleanup tag to image: %v\n", err)
	}

	if verbose {
		time.Sleep(1 * time.Second)
		fmt.Printf("Backing up instance Id: %s named %s completed. New AMI: %s\n", abkupInstance.InstanceId, abkupInstance.Name, createImageResp.ImageId)
	}

}

func main() {
	// pointers to objects we use to talk to AWS
	var e *ec2.EC2

	// storage for commandline args
	var regionName, awsKey, awsSecret, awsRegion string
	var autoFlag bool
	var bkupId string

	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key")
	flag.BoolVar(&verbose, "v", false, "Produce verbose output")

	flag.BoolVar(&autoFlag, "a", false, "In auto mode snapshot any instance with an autobkup tag")
	flag.StringVar(&bkupId, "i", "", "Instance id to be backed up")
	flag.Parse()

	// make sure we are in auto mode or an ami id has been provided
	if !autoFlag && len(bkupId) == 0 {
		fmt.Printf("No instance details provided. Please provide an instance id to snapshot\nor enable auto mode to snapshot all tagged instances.\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// load the AWS credentials from the environment or from the standard file
	loadAWSCredentials(awsKey, awsSecret, regionName)

	// pull the region from the environment
	awsRegion = os.Getenv("AWS_ACCESS_REGION")

	// Pull the access details from environment variables
	// AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Fatalf("Error unable to find Access and/or Secret key\n")
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
		log.Fatalf("Error - Sorry I can not find the url endpoint for region %s\n", awsRegion)
	}

	// load the struct that has details on all instances to be snapshotted
	bkupInstances := getBkupInstances(e, bkupId)

	// now we have the slice of instances to be backed up we can create the AMI then tag them

	var wg sync.WaitGroup

	// rate limit the AWS requests to max 3 per second
	rl := rate.New(3, time.Second)

	for instance := range bkupInstances {

		rl.Wait()

		// Increment the WaitGroup counter.
		wg.Add(1)

		// Launch a goroutine to fetch the URL.
		go func(e *ec2.EC2, abkupInstance ec2.CreateImage) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// snapshot the instance.
			ssInstance(e, &abkupInstance)
		}(e, bkupInstances[instance])

	}

	// Wait for all Amazon requests to complete.
	wg.Wait()

	if verbose {
		fmt.Printf("All done.\n")

	}

}
