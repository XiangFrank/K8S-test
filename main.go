package main

import (
	"bufio"
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
	"os"
)

func main() {
	namespace := "yeda-test"
	jobId := "test"
	imageId := "yeda:v1"
	KubeConfig := "/home/yeda/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", KubeConfig)
	if err != nil {
		fmt.Println("error 1")
		fmt.Println(err.Error())
		//TODO: handle error
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("error 2")
		fmt.Println(err.Error())
		//TODO: handle error
	}

	fmt.Println("Create client.")

	prompt()

	depClient := client.AppsV1().Deployments(namespace)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobId + "-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": jobId,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": jobId,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name: jobId + "-container",
							Image: imageId,
							Ports: []apiv1.ContainerPort{
								{
									Name: "http",
									Protocol: apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
							Command: []string{
								"/bin/bash",
								"-c",
								"--",
							},
							Args: []string{
								"while true; do sleep 30; done;",
							},
						},

					},
				},
			},
		},
	}


	result, err := depClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("error 3")
		fmt.Println(err.Error())
		//TODO: handle error
	}

	fmt.Println(result.UID)

	prompt()

	service, err := client.CoreV1().Services(namespace).Create(context.TODO(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobId + "-service",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "controller",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type: 0,
						IntVal: 32730,
					},
				},
			},
			Selector: map[string]string{
				"app": jobId,
			},
			ClusterIP: "",
		},
	}, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("create service failed: " + err.Error())
	}

	fmt.Println("service clusterIP is " + service.Spec.ClusterIP)

	//fmt.Println("service external IP is " + service.Spec.ExternalIPs[0])

	prompt()

	listOptions := metav1.ListOptions{LabelSelector: labels.Set(service.Spec.Selector).AsSelector().String()}
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	for _, pod := range pods.Items{
		fmt.Fprintln(os.Stdout,"pod is :", pod.Name)
		fmt.Println(pod.Status.PodIP)
	}

	prompt()

	deletePolicy := metav1.DeletePropagationBackground
	if err := depClient.Delete(context.TODO(), jobId+"-deployment", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil{
		fmt.Println("error 4")
		fmt.Println(err)
		//TODO: handle error
	}

	if err := client.CoreV1().Services(namespace).Delete(context.TODO(), jobId + "-service", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil{
		//TODO: handle error
		fmt.Println("error 5")
		fmt.Println(err)
	}
	fmt.Println("Deleted deployment")

}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}