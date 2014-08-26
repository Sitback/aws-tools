# Sitback AWS Tools

This repo contains code that Sitback has created for its technical operations.
The code was created to ease the daily work burden and also to upskill
with Go Language. Enjoy

At the moment there are a number of different Github repos for AWS. We have
decided to use https://github.com/goamz/goamz.


The following applications are available at the moment:

## aws-autostop

This simple program will search for any instances that have a tag of autostop
and are in state running. If any are found then the instance is stopped.
This program can be called from a cronjob to shutdown instances
that do not need to be running 24x7 such as staging servers.



## aws-check-reserved-instances

This simple program will read your AWS credentials from the environment or
your ${HOME}/.aws/config file then search for any reserved instances and
display information about them. Use the -d or -e to select instances that
are going to expire soon or have recently expired.
The following will display reserved instances due to expire in the next
30 days.

Example: aws-check-reserved-instances -d 30

NOTE: This does not use goamz/goamz library yet as that library does not
yet support reserved instances.



## aws-describe-as-groups

This simple program will display the names of any auto scale
groups you have configured in the AWS region.



## aws-describe-asg

This simple program will display the private ip addresses of
any instances in the auto scaling group. Good to get the internal ip
addresses if you need to connect to all servers in the group.




## aws-describe-instances

This simple program will display basic info about the instances in the region.



## aws-ss-instance
## aws-ami-cleanup

These two programs will allow you to create an AMI for an instance and then
automatically cleanup old AMI's after a few days. This is useful if you
want to create a snapshot of your instances.
From cron first run aws-ss-instance -
14 03 * * * username /usr/local/bin/aws-ss-instance -a

This will snapshot all instances hat have a tag named autobkup in the account.

24 03 * * * username /usr/local/bin/aws-ami-cleanup -a 2

This will remove any old AMI's and associated snapshots that are older than 2 days.



