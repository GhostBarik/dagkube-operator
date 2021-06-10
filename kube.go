package main

import (
	"flag"
	"fmt"
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
	Metadata TaskMetadata
}

func (n *KubeTask) GetId() TaskId {
	return n.Id
}

func (n *KubeTask) Run() error {
	fmt.Printf("task[%v]: started\n", n.Metadata.Name)
	time.Sleep(1 * time.Second)
	fmt.Printf("task[%v]: finished\n", n.Metadata.Name)
	return nil
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

func CreateJobDefinition(jobName string, containerName string) batchv1.Job {
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec {
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{{
						Name: containerName,
						Image: "acrdagkube.azurecr.io/dagkube-poc:v0.1.0",
						Args: []string{ "30", "0.8" },
						ImagePullPolicy: "Always",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu": resource.MustParse("300m"),
								"memory": resource.MustParse("256Mi"),
							},
						},
					}},
				},
			},
		},
	}
}
