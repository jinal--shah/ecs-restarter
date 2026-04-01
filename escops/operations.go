package ecsops

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func RestartService(ctx context.Context, client *ecs.Client, cluster, service string) error {
	_, err := client.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:            &cluster,
		Service:            &service,
		ForceNewDeployment: true,
	})
	return err
}

func ScaleService(ctx context.Context, client *ecs.Client, cluster, service string, desired int32) error {
	_, err := client.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:            &cluster,
		Service:            &service,
		DesiredCount:       &desired,
		ForceNewDeployment: true,
	})
	return err
}

func WaitForStable(ctx context.Context, client *ecs.Client, cluster, service string) error {
	fmt.Println("Waiting for service to become stable...")

	for {
		out, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  &cluster,
			Services: []string{service},
		})
		if err != nil {
			return err
		}

		if len(out.Services) == 0 {
			return fmt.Errorf("service not found")
		}

		svc := out.Services[0]

		running := svc.RunningCount
		desired := svc.DesiredCount

		fmt.Printf("Progress: %d/%d tasks running\n", running, desired)

		// Stability condition
		if running == desired && len(svc.Deployments) == 1 {
			fmt.Println("Service is stable ✅")
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}
