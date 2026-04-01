package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"ecs-restarter/ui"
	"ecs-restarter/escops"
)

type result struct {
	cluster string
	service string
	running int32
	desired int32
	err     error
}

func main() {
	region := flag.String("region", "eu-west-1", "AWS region")
	filter := flag.String("filter", "", "Substring to match ECS service names")
	workers := flag.Int("workers", 5, "Number of concurrent workers")
	flag.Parse()

	if *filter == "" {
		log.Fatal("You must provide a -filter value")
	}

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(*region),
	)
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	client := ecs.NewFromConfig(cfg)

	// ... get all ECS clusters for the region
	var clusters []string
	clusterPaginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{})

	for clusterPaginator.HasMorePages() {
		page, err := clusterPaginator.NextPage(ctx)
		if err != nil {
			log.Fatalf("error listing clusters: %v", err)
		}
		clusters = append(clusters, page.ClusterArns...)
	}

	clusterCh := make(chan string)
	resultCh := make(chan result)

	var wg sync.WaitGroup

	// ... workers process clusters
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for clusterArn := range clusterCh {
				processCluster(ctx, client, clusterArn, *filter, resultCh)
			}
		}()
	}

	// ... ecs clusters
	go func() {
		for _, c := range clusters {
			clusterCh <- c
		}
		close(clusterCh)
	}()

	// ... wait for results to come in.
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// ... find matching ecs-service
	var results []result

	for res := range resultCh {
		if res.err != nil {
			log.Printf("error: %v", res.err)
			continue
		}
		results = append(results, res)
	}

	if len(results) == 0 {
		log.Println("No matching services found")
		return
	}

	// ... gui options from results
	options := make([]ui.Option[result], 0, len(results))

	for _, r := range results {
		options = append(options, ui.Option[result]{
			Label: fmt.Sprintf(
				"%s | %s | %d/%d",
				r.cluster,
				r.service,
				r.running,
				r.desired,
			),
			Data: r,
		})
	}

	// ... let user select
	selected, err := ui.SelectOption(options)
	if err != nil {
		log.Fatalf("selection failed: %v", err)
	}

	// ... get cluster and service from selection
	cluster := selected.Data.cluster
	service := selected.Data.service
	currentDesired := selected.Data.desired

	// ... let user choose action
	action, err := ui.SelectAction()
	if err != nil {
		log.Fatal(err)
	}

	switch action {

		case ui.RestartService:
			if !ui.ConfirmRestart(service) {
				fmt.Println("Cancelled.")
				return
			}

			fmt.Println("Restarting service...")

			err := ecsops.RestartService(ctx, client, cluster, service)
			if err != nil {
				log.Fatalf("failed to restart: %v", err)
			}

			err = ecsops.WaitForStable(ctx, client, cluster, service)
			if err != nil {
				log.Fatalf("wait failed: %v", err)
			}

			fmt.Println("Restart completed.")

		case ui.ScaleService:
			newDesired, err := ui.PromptScale(currentDesired)
			if err != nil {
				log.Fatalf("invalid input: %v", err)
			}

			if newDesired == currentDesired {
				fmt.Println("Desired count unchanged. Exiting.")
				return
			}

			fmt.Printf("Scaling service to %d tasks...\n", newDesired)

			err = ecsops.ScaleService(ctx, client, cluster, service, newDesired)
			if err != nil {
				log.Fatalf("scale failed: %v", err)
			}

			err = ecsops.WaitForStable(ctx, client, cluster, service)
			if err != nil {
				log.Fatalf("wait failed: %v", err)
			}

			fmt.Println("Scaling successful.")
	}
}

// processCluster() - return matching ecs-service details for a given cluster from its arn and the filter arg
func processCluster(ctx context.Context, client *ecs.Client, clusterArn string, filter string, resultCh chan<- result) {
	clusterName := clusterArn[strings.LastIndex(clusterArn, "/")+1:]

	servicePaginator := ecs.NewListServicesPaginator(client, &ecs.ListServicesInput{
		Cluster: &clusterArn,
	})

	var matching []string

	for servicePaginator.HasMorePages() {
		page, err := servicePaginator.NextPage(ctx)
		if err != nil {
			resultCh <- result{err: fmt.Errorf("list services (%s): %w", clusterName, err)}
			return
		}

		for _, arn := range page.ServiceArns {
			if strings.Contains(arn, filter) {
				matching = append(matching, arn)
			}
		}
	}

	if len(matching) == 0 {
		return
	}

	// batch limit hardcoded to 10
	for i := 0; i < len(matching); i += 10 {
		end := i + 10
		if end > len(matching) {
			end = len(matching)
		}

		batch := matching[i:end]

		out, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  &clusterArn,
			Services: batch,
		})
		if err != nil {
			resultCh <- result{err: fmt.Errorf("describe services (%s): %w", clusterName, err)}
			return
		}

		for _, svc := range out.Services {
			resultCh <- result{
				cluster: clusterName,
				service: *svc.ServiceName,
				running: svc.RunningCount,
				desired: svc.DesiredCount,
			}
		}
	}
}
