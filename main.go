package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	csvfile, err := os.Open("metrics.csv")
	if err != nil {
		panic(err.Error())
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = 4 // Deployment name, Namespace, CPU limit, and Memory limit

	ctx := context.TODO() // Using TODO context for now, it's similar to context.Background()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue
		}

		deploymentName := record[0]
		namespace := record[1]
		cpuLimit := record[2]
		memoryLimit := record[3]

		deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Failed to get deployment %s in namespace %s: %v\n", deploymentName, namespace, err)
			continue
		}

		// Set resource limits
		deployment.Spec.Template.Spec.Containers[0].Resources.Limits = v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuLimit),
			v1.ResourceMemory: resource.MustParse(memoryLimit),
		}

		_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Failed to update deployment %s in namespace %s: %v\n", deploymentName, namespace, err)
		} else {
			fmt.Printf("Successfully updated deployment %s in namespace %s\n", deploymentName, namespace)
		}
	}
}
