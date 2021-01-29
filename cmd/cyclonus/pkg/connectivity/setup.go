package connectivity

import (
	connectivitykube "github.com/mattfenwick/cyclonus/pkg/connectivity/kube"
	"github.com/mattfenwick/cyclonus/pkg/connectivity/synthetic"
	"github.com/mattfenwick/cyclonus/pkg/kube"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"time"
)

func SetupClusterTODODelete(kubernetes *kube.Kubernetes, namespaces []string, pods []string, port int, protocol v1.Protocol) (*connectivitykube.Resources, *synthetic.Resources, error) {
	kubeResources := connectivitykube.NewDefaultResources(namespaces, pods, []int{port}, []v1.Protocol{protocol})

	err := kubeResources.CreateResourcesInKube(kubernetes)
	if err != nil {
		return nil, nil, err
	}

	err = waitForPodsReadyTODODelete(kubernetes, namespaces, pods, 60)
	if err != nil {
		return nil, nil, err
	}

	podList, err := kubernetes.GetPodsInNamespaces(namespaces)
	if err != nil {
		return nil, nil, err
	}
	var syntheticPods []*synthetic.Pod
	for _, pod := range podList {
		ip := pod.Status.PodIP
		if ip == "" {
			return nil, nil, errors.Errorf("no ip found for pod %s/%s", pod.Namespace, pod.Name)
		}
		syntheticPods = append(syntheticPods, &synthetic.Pod{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Labels:    pod.Labels,
			IP:        ip,
		})
		log.Infof("ip for pod %s/%s: %s", pod.Namespace, pod.Name, ip)
	}

	syntheticResources, err := synthetic.NewResources(kubeResources.Namespaces, syntheticPods)
	if err != nil {
		return nil, nil, err
	}

	return kubeResources, syntheticResources, nil
}

func waitForPodsReadyTODODelete(kubernetes *kube.Kubernetes, namespaces []string, pods []string, timeoutSeconds int) error {
	sleep := 5
	for i := 0; i < timeoutSeconds; i += sleep {
		podList, err := kubernetes.GetPodsInNamespaces(namespaces)
		if err != nil {
			return err
		}

		ready := 0
		for _, pod := range podList {
			if pod.Status.Phase == "Running" && pod.Status.PodIP != "" {
				ready++
			}
		}
		if ready == len(namespaces)*len(pods) {
			return nil
		}

		log.Infof("waiting for pods to be running and have IP addresses")
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return errors.Errorf("pods not ready")
}

func SetupCluster(kubernetes *kube.Kubernetes, kubeResources *connectivitykube.Resources) error {
	err := kubeResources.CreateResourcesInKube(kubernetes)
	if err != nil {
		return err
	}

	err = waitForPodsReady(kubernetes, kubeResources, 60)
	if err != nil {
		return err
	}
	return nil
}

func GetSyntheticResources(kubernetes *kube.Kubernetes, kubeResources *connectivitykube.Resources) (*synthetic.Resources, error) {
	podList, err := kubernetes.GetPodsInNamespaces(kubeResources.NamespacesSlice())
	if err != nil {
		return nil, err
	}
	var syntheticPods []*synthetic.Pod
	for _, pod := range podList {
		ip := pod.Status.PodIP
		if ip == "" {
			return nil, errors.Errorf("no ip found for pod %s/%s", pod.Namespace, pod.Name)
		}
		var containers []*synthetic.Container
		for _, kubeCont := range pod.Spec.Containers {
			if len(kubeCont.Ports) != 1 {
				return nil, errors.Errorf("expected 1 port on kube container, found %d", len(kubeCont.Ports))
			}
			kubePort := kubeCont.Ports[0]
			containers = append(containers, &synthetic.Container{
				Port:     int(kubePort.ContainerPort),
				Protocol: kubePort.Protocol,
			})
		}
		syntheticPods = append(syntheticPods, &synthetic.Pod{
			Namespace:  pod.Namespace,
			Name:       pod.Name,
			Labels:     pod.Labels,
			IP:         ip,
			Containers: containers,
		})
		log.Infof("ip for pod %s/%s: %s", pod.Namespace, pod.Name, ip)
	}

	syntheticResources, err := synthetic.NewResources(kubeResources.Namespaces, syntheticPods)
	if err != nil {
		return nil, err
	}

	return syntheticResources, nil
}

func waitForPodsReady(kubernetes *kube.Kubernetes, kubeResources *connectivitykube.Resources, timeoutSeconds int) error {
	sleep := 5
	for i := 0; i < timeoutSeconds; i += sleep {
		podList, err := kubernetes.GetPodsInNamespaces(kubeResources.NamespacesSlice())
		if err != nil {
			return err
		}

		ready := 0
		for _, pod := range podList {
			if pod.Status.Phase == "Running" && pod.Status.PodIP != "" {
				ready++
			}
		}
		if ready == len(kubeResources.Pods) {
			return nil
		}

		log.Infof("waiting for pods to be running and have IP addresses")
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return errors.Errorf("pods not ready")
}
