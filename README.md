aws-tools
=========

Sitback AWS Tools

This repo contains code that Sitback has created for its technical operations.
The code was created to ease the daily work burden and also to upskill
with Go Language. Enjoy




The following applications are available at the moment:

aws-autostop
This simple program will read your AWS credentials from the environment or
your ${HOME}/.aws/config file then search for any instances that have a
tag of autostop and are in state running. If any are found then the instance
is stopped. This program can be called from a cronjob to shutdown instances
that do not need to be running 24x7 such as staging servers.



aws-check-reserved-instances
This simple program will read your AWS credentials from the environment or
your ${HOME}/.aws/config file then search for any reserved instances and
display information about them. Use the -d or -e to select instances that
are going to expire soon or have recently expired.

The following will display reserved instances due to expire in the next
30 days.
Example: aws-check-reserved-instances -d 30



aws-describe-as-groups
This simple program will read your AWS credentials from the environment or
your ${HOME}/.aws/config file then display the names of any auto scale
groups you have configured in the region.



aws-describe-instances
This simple program will read your AWS credentials from the environment or
your ${HOME}/.aws/config file then display basic info about the instances
in the region


