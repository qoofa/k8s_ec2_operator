package controller

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	computev1 "github.com/qoofa/ec2-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func deleteInstance(ctx context.Context, ec2Instance *computev1.Ec2instance) (bool, error) {
	l := log.FromContext(ctx)

	instanceId := ec2Instance.Status.InstanceID

	l.Info("Deleting Instance : ", "instance Id", instanceId)

	client := awsClient(ec2Instance.Spec.Region)

	terminateResult, err := client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		l.Error(err, "failed to terminate ec2 instance")
		return false, err
	}

	l.Info("Instance termination initiated",
		"instanceId", instanceId,
		"currentState", terminateResult.TerminatingInstances[0].CurrentState.Name)

	waiter := ec2.NewInstanceTerminatedWaiter(client)

	maxWaitTime := time.Minute * 5
	waitPayload := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}

	l.Info("Waiting for instance deletion to be completed",
		"instanceId", instanceId,
		"maxWaitTime", maxWaitTime)

	err = waiter.Wait(ctx, waitPayload, maxWaitTime)
	if err != nil {
		l.Error(err, "failed wihle waiting for instance termination",
			"instanceId", instanceId)
		return false, err
	}

	l.Info("Ec2 instance successfully terminated", "instanceId", instanceId)
	return true, nil
}
