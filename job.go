package main

import (
	"context"
	"errors"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	b1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"time"
)

type KubeTask struct {

	// generated during graph parsing / creation
	Id TaskId
	// task metadata (e.g. which image to run, arguments etc.)
	Metadata KubeTaskMetadata

	// kubernetes job client
	jobClient b1.JobInterface
	// kubernetes job definition
	jobDefinition batchv1.Job
}

type KubeTaskMetadata struct {
	Name            string
	Image           string   // url of the docker image (with tag)
	Args            []string // list of string arguments to pass
	NumberOfRetries int
	// TODO: add additional properties (e.g. container limits)
}

func (n *KubeTask) GetId() TaskId {
	return n.Id
}

func (n *KubeTask) Run() error {

	image := "acrdagkube.azurecr.io/dagkube-poc:v0.1.0"
	jobBaseName := "dagkube"
	containerName := "dagkube-job-container"
	retries := 3
	jobName := jobBaseName + "-" + n.Metadata.Name

	jobDefinition := CreateJobDefinition(jobName, image, containerName, int32(retries))

	// create the job
	result, err := n.jobClient.Create(context.TODO(), &jobDefinition, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("KubeTask(%v): Cannot create job, %v", jobName, err)
		return err
	}

	fmt.Printf(
		"KubeTask(%v): created job - %q.\n",
		jobName, result.GetObjectMeta().GetName(),
	)

	//var numberOfCompletionToSucceed int32 = 1

	// watch for the job status
	for {

		result, _ = n.jobClient.Get(context.TODO(), jobName, metav1.GetOptions{})

		successCheck := result.Status.Succeeded
		fmt.Printf("status: %v\n", successCheck)

		fmt.Printf(
			"KubeTask(%v): job status: Succeeded(%v), Failed(%v), Active(%v)\n",
			jobName,
			successCheck,
			result.Status.Failed == 1,
			result.Status.Active == 1,
		)

		if successCheck > 0 {
			fmt.Printf(
				"KubeTask(%v): finished with success %q.\n",
				jobName, result.GetObjectMeta().GetName(),
			)
			return nil
		}
		if result.Status.Failed == 1 {
			fmt.Printf(
				"KubeTask(%v): finished with error %q.\n",
				jobName, result.GetObjectMeta().GetName(),
			)
			return errors.New("task failed")
		}
		// wait 1s until next check
		time.Sleep(1 * time.Second)
	}
}
