package main

import (
	"errors"
	"flag"
	"fmt"
	"context"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	b1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

type KubeTask struct {
	// generated during graph parsing, do not change!
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

	image := "acrdagkube.azurecr.io/dagkube-poc:v0.1.1"
	jobBaseName := "dagkube"
	containerName := "dagkube-job-container"
	retries := int32(2)
	jobName := jobBaseName + "-" + n.Metadata.Name
	params := []string{"1", "0.1"}

	jobDefinition := CreateJobDefinition(jobName, image, containerName, params, int32(retries))

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

	// watch for the job status
	for {

		result, _ = n.jobClient.Get(context.TODO(), jobName, metav1.GetOptions{})

		numberOfSuccesses := result.Status.Succeeded
		numberOfFailures := result.Status.Failed

		fmt.Printf("status: %v\n", numberOfSuccesses)

		fmt.Printf(
			"KubeTask(%v): job status: Successes(%v), Failures(%v)\n",
			jobName,
			numberOfSuccesses,
			numberOfFailures,
		)

		if numberOfSuccesses > 0 {
			fmt.Printf(
				"KubeTask(%v): finished with success %q.\n",
				jobName, result.GetObjectMeta().GetName(),
			)
			return nil
		}

		// TODO: check job success value instead??? (i.e. do not rely on our number of restarts)
		// TODO: if there is some parameter, that just says "job has failde after all the reties"?
		if numberOfFailures > retries {
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

func CreateJobClient(namespace string) b1.JobInterface {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String(
			"kubeconfig", "",
			"absolute path to the kubeconfig file",
		)
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset.BatchV1().Jobs(namespace)
}

func CreateJobDefinition(jobName string, image string, containerName string, params []string, retries int32) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &retries,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{{
						Name:            containerName,
						Image:           image,
						Args:            params,
						ImagePullPolicy: "IfNotPresent",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse("300m"),
								"memory": resource.MustParse("256Mi"),
							},
						},
					}},
				},
			},
		},
	}
}
