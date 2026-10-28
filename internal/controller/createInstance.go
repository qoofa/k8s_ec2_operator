package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	computev1 "github.com/qoofa/ec2-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func createEc2Instance(data *computev1.Ec2instance) (createdInstanceInfo *computev1.CreatedInstanceInfo, err error) {
	l := log.Log.WithName("CreateEc2Instance")

	l.Info("Creating EC2 instance",
		"amiId", data.Spec.AMIId,
		"instanceType", data.Spec.InstanceType,
		"region", data.Spec.Region)

	client := awsClient(data.Spec.Region)

	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(data.Spec.AMIId),
		InstanceType: ec2types.InstanceType(data.Spec.InstanceType),
		KeyName:      aws.String(data.Spec.KeyPair),
		SubnetId:     aws.String(data.Spec.Subnet),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	result, err := client.RunInstances(context.TODO(), runInput)
	if err != nil {
		l.Error(err, "Failed to run EC2 instance")
		return nil, fmt.Errorf("failed to create ec2 instance %w", err)
	}

	if len(result.Instances) == 0 {
		l.Error(nil, "RunInstances returned no instance")
		return nil, fmt.Errorf("no instance returned by RunInstance api")
	}

	inst := result.Instances[0]
	l.Info("Launched EC2 instance, waiting for running state", "instanceId", *inst.InstanceId)

	runWaiter := ec2.NewInstanceRunningWaiter(client)
	maxWaitTime := 3 * time.Minute

	describeInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{*inst.InstanceId},
	}

	err = runWaiter.Wait(context.TODO(), describeInput, maxWaitTime)
	if err != nil {
		l.Error(err, "Failed to wait for instance to be running")
		return nil, fmt.Errorf("failed to wait for instance to be running: %w", err)
	}

	describeResult, err := client.DescribeInstances(context.TODO(), describeInput)
	if err != nil {
		l.Error(err, "Failed to describe EC2 instance")
		return nil, fmt.Errorf("failed to describe EC2 instance: %w", err)
	}

	fmt.Println("Describe result", "public ip", *describeResult.Reservations[0].Instances[0].PublicDnsName, "state", describeResult.Reservations[0].Instances[0].State.Name)

	fmt.Printf("Private IP of the instance: %v", derefString(inst.PrivateIpAddress))
	fmt.Printf("State of the instance: %v", describeResult.Reservations[0].Instances[0].State.Name)
	fmt.Printf("Private DNS of the instance: %v", derefString(inst.PrivateDnsName))
	fmt.Printf("Instance ID of the instance: %v", derefString(inst.InstanceId))
	fmt.Println("Instance Type of the instance: ", inst.InstanceType)
	fmt.Printf("Image ID of the instance: %v", derefString(inst.ImageId))
	fmt.Printf("Key Name of the instance: %v", derefString(inst.KeyName))

	instance := describeResult.Reservations[0].Instances[0]
	createdInstanceInfo = &computev1.CreatedInstanceInfo{
		InstanceID: *inst.InstanceId,
		State:      string(instance.State.Name),
		PublicIP:   derefString(instance.PublicIpAddress),
		PrivateIP:  derefString(instance.PrivateIpAddress),
		PublicDNS:  derefString(instance.PublicDnsName),
		PrivateDNS: derefString(instance.PrivateDnsName),
	}

	l.Info("EC2 instance is running",
		"instanceId", createdInstanceInfo.InstanceID,
		"state", createdInstanceInfo.State,
		"publicIP", createdInstanceInfo.PublicIP)

	return createdInstanceInfo, nil
}

func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return "<nil>"
}
