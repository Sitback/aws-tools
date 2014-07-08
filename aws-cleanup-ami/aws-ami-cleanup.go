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
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/gombadi/goamz/ec2"
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

func cleanupAMI(e *ec2.EC2, cleanupImage ec2.Image) {
	if verbose {
		fmt.Printf("Info - Deregistering AMI: %s\n", cleanupImage.Id)
	}
	// looks like we have found an ami to cleanup
	deregisterImageResp, err := e.DeregisterImage(cleanupImage.Id)

	if err != nil || deregisterImageResp.Response != true {
		fmt.Printf("\nError deregistering the image %s...\n%v\n\n", cleanupImage.Id, err)
		return
	}

	if verbose {
		fmt.Printf("Image has been deregistered and now waiting 2 seconds for AWS to release the snapshots from the AMI\n")
	}

	// after the image is deregistered then you can delete the snapshots it used
	time.Sleep(2 * time.Second)

	for blockDevice := range cleanupImage.BlockDevices {
		if len(cleanupImage.BlockDevices[blockDevice].SnapshotId) > 0 {
			if verbose {
				fmt.Printf("Info - Deleting associated snapshot: %s from ami: %s\n", cleanupImage.BlockDevices[blockDevice].SnapshotId, cleanupImage.Id)
			}

			_, err := e.DeleteSnapshots(cleanupImage.BlockDevices[blockDevice].SnapshotId)
			if err != nil {
				fmt.Printf("\nError deleting the snapshots %s.\n%v\n\n", cleanupImage.BlockDevices[blockDevice].SnapshotId, err)
			}
		}
	}
	if verbose {
		fmt.Printf("Snapshots have been deleted.\n")
	}
}

func main() {

	// pointers to objects we use to talk to AWS
	var e *ec2.EC2
	var imagesResp *ec2.ImagesResp

	// storage for commandline args
	var regionName, awsKey, awsSecret, awsRegion string
	var autoDays int
	var amiId string

	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key")
	flag.BoolVar(&verbose, "v", false, "Produce verbose output")

	flag.IntVar(&autoDays, "a", 0, "In auto cleanup mode cleanup any AMI's older than this number of days")
	flag.StringVar(&amiId, "i", "", "AMI Id to be deleted")
	flag.Parse()

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

	// make sure we are in auto mode or an ami id has been provided
	if autoDays == 0 && len(amiId) == 0 {
		fmt.Printf("No ami details provided. Please provide an ami-id to cleanup\nor enable auto cleanup mode and specify a number of days.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// config for auto mode
	if autoDays > 0 {

		options := ec2.DescribeImagesOptions{
			Owner: "self", // Only search my own images
			// ExecutableBy: "all"
		}

		// auto mode search for ami's to cleanup
		filter := ec2.NewFilter()
		filter.Add("tag-key", "autocleanup")

		imagesResp, err = e.Images(nil, filter, &options)
		if err != nil {
			log.Fatalf("\nError getting the Image details for images.\n%v\n", err)
		}
	} else {
		// single ami manual mode
		amiIdsSlice := []string{amiId}

		options := ec2.DescribeImagesOptions{}

		imagesResp, err = e.Images(amiIdsSlice, nil, &options)
		if err != nil {
			log.Fatalf("\nError getting the Image details for image %s...\n%v\n", amiId, err)
		}
	}

	if len(imagesResp.Images) == 0 {
		if verbose {
			fmt.Printf("No images found to cleanup. Exiting\n")
		}
		os.Exit(0)
	}

	// extract the instanceId with autostop tags and state running
	for image := range imagesResp.Images {

		// The returned Images from AWS should only be the ones with autocleanup but lets check anyway
		// and only delete if the days have passed

		for tag := range imagesResp.Images[image].Tags {
			if imagesResp.Images[image].Tags[tag].Key == "autocleanup" {
				// check if time is up for this AMI

				// extract the time this AMI was created
				t, _ := time.Parse(time.RFC3339, imagesResp.Images[image].Tags[tag].Value)
				elapsed := time.Since(t)

				if (autoDays * 86400) < int(elapsed.Seconds()) {
					// deregister the AMI and delete associated snapshots
					cleanupAMI(e, imagesResp.Images[image])
					//fmt.Printf("\nWould have cleaned up AMI: %s\n\n", imagesResp.Images[image].Id)
				}
			}
		}

	}

	if verbose {
		fmt.Printf("All done.\n")
	}

}
