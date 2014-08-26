/*
	This program will display information about reserved instances and when they expire.
	As Amazon will not tell you when a reserved instance has expired and let it continue
	as on demand it is a good idea to keep an eye on reserved instances

*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gombadi/goamz/aws"
	"github.com/gombadi/goamz/ec2"
	"github.com/gombadi/go-ini"
)

func main() {

	// pointers to objects we use to talk to AWS
	var e *ec2.EC2
	// storage for commandline args
	var regionName, awsKey, awsSecret string
	var expireDays, retireDays int
	var ok bool

	flag.StringVar(&regionName, "r", "xxxx", "AWS Region to send request. Read from ${HOME}/.aws/config if not provided")
	flag.StringVar(&awsKey, "k", "xxxx", "AWS Access Key.Read from ${HOME}/.aws/config if not provided")
	flag.StringVar(&awsSecret, "s", "xxxx", "AWS Secret key. Read from ${HOME}/.aws/config if not provided")
	// flags for reserved instances
	flag.IntVar(&expireDays, "d", 0, "Number of days till expire")
	flag.IntVar(&retireDays, "e", 0, "Retired in last number of days")
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

	instanceResp, err := e.DescribeReservedInstances(nil, nil)
	if err != nil {
		fmt.Println("\nBother, things are not looking good getting the Reserved Instance details...\n")
		panic(err)
	}

	switch {
	case expireDays > 0:
		fmt.Printf("Displaying reserved instances that will expire within the next %d days\n\n", expireDays)
		// if expireDays provided then ignore retireDays
		retireDays = 0
	case retireDays > 0:
		fmt.Printf("Displaying reserved instances that have expired within the last %d days\n\n", retireDays)
	}

	// extract the private ip address from the instance struct stored in the reservation
	for reservedInstance := range instanceResp.ReservedInstances {

		// extract the time this reserved instance expires
		t, _ := time.Parse(time.RFC3339, instanceResp.ReservedInstances[reservedInstance].End)

		elapsed := time.Since(t)

		if expireDays > 0 {
			if (int(elapsed.Hours()/24) > (0 - expireDays)) && instanceResp.ReservedInstances[reservedInstance].State == "active" {
				// Display info on instances that will expire soon
				fmt.Printf("Reserved_instance: %s\tOffer_Type: %s\tState: %s\nNum_of_instances: %d\tInstance_Type: %s\tAV-Zone: %s\tExpire: %s\n\n",
					instanceResp.ReservedInstances[reservedInstance].ReservedInstanceId,
					instanceResp.ReservedInstances[reservedInstance].OfferingType,
					instanceResp.ReservedInstances[reservedInstance].State,
					instanceResp.ReservedInstances[reservedInstance].InstanceCount,
					instanceResp.ReservedInstances[reservedInstance].InstanceType,
					instanceResp.ReservedInstances[reservedInstance].AvailabilityZone,
					instanceResp.ReservedInstances[reservedInstance].End)
			}
		}

		if retireDays > 0 {
			if (retireDays > int(elapsed.Hours()/24)) && instanceResp.ReservedInstances[reservedInstance].State == "retired" {
				// Display info on instances that will expire soon
				fmt.Printf("Reserved_instance: %s\tOffer_Type: %s\tState: %s\nNum_of_instances: %d\tInstance_Type: %s\tAV-Zone: %s\tExpire: %s\n\n",
					instanceResp.ReservedInstances[reservedInstance].ReservedInstanceId,
					instanceResp.ReservedInstances[reservedInstance].OfferingType,
					instanceResp.ReservedInstances[reservedInstance].State,
					instanceResp.ReservedInstances[reservedInstance].InstanceCount,
					instanceResp.ReservedInstances[reservedInstance].InstanceType,
					instanceResp.ReservedInstances[reservedInstance].AvailabilityZone,
					instanceResp.ReservedInstances[reservedInstance].End)
			}
		}



	}
}
