/*
Copyright 2024 Abhijeet Rokade.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/sigsegv1989/flipper-operator/test/utils"
)

const namespace = "flipper-operator-system"

var kindClusterName = "e2e-test"
var KindClusterConfig = "/home/abhijeet/.kube/" + kindClusterName + ".kubeconfig"

var _ = Describe("controller", Ordered, func() {

	var clientset *kubernetes.Clientset
	var dynamicClient dynamic.Interface

	BeforeAll(func() {
		By("exporting KIND_CLUSTER environment variable")
		os.Setenv("KIND_CLUSTER", kindClusterName)

		By("checking if the Kind cluster for e2e tests exists")
		cmd := exec.Command("kind", "get", "clusters")
		output, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())

		if !strings.Contains(string(output), kindClusterName) {
			By("creating a Kind cluster for e2e tests")
			cmd = exec.Command("kind", "create", "cluster", "--name", kindClusterName, "--kubeconfig", KindClusterConfig)
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
		} else {
			By("kind cluster already exists, skipping creation")
		}

		By("exporting KUBECONFIG environment variable")
		os.Setenv("KUBECONFIG", KindClusterConfig)

		By("load kubeconfig")
		configPath := os.ExpandEnv(KindClusterConfig)
		config, err := clientcmd.BuildConfigFromFlags("", configPath)
		Expect(err).NotTo(HaveOccurred())

		By("create Kubernetes client")
		clientset, err = kubernetes.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())

		By("create dynamic client")
		dynamicClient, err = dynamic.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())

		/*
			By("installing prometheus operator")
			Expect(utils.InstallPrometheusOperator()).To(Succeed())

			By("installing the cert-manager")
			Expect(utils.InstallCertManager()).To(Succeed())
		*/

		By("creating manager namespace")
		cmd = exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {
		/*
			By("uninstalling the Prometheus manager bundle")
			utils.UninstallPrometheusOperator()

			By("uninstalling the cert-manager bundle")
			utils.UninstallCertManager()
		*/

		By("removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)

		By("deleting the Kind cluster")
		cmd = exec.Command("kind", "delete", "cluster", "--name", kindClusterName)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error

			// projectimage stores the name of the image used in the example
			var projectimage = "example.com/flipper-operator:v0.0.1"

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			err = utils.LoadImageToKindClusterWithName(projectimage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

		})
	})

	// Helper function to apply YAML manifest
	applyManifest := func(path string) {
		cmd := exec.Command("kubectl", "apply", "-f", path)
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", KindClusterConfig))

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to apply manifest %s, error: %v, output: %s\n", path, err, string(output))
			Expect(err).NotTo(HaveOccurred())
		}
	}

	// Helper function to check if deployment pods have restarted
	checkDeploymentPodsRestarted := func(deploymentName, namespace string) {
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			// Check if the annotation exists and is not empty
			restartedAtExists := deployment.Annotations["kubectl.kubernetes.io/restartedAt"] != ""
			restartedBy := deployment.Annotations["kubectl.kubernetes.io/restartedBy"] == "flipper-operator"
			restartedByCRExists := deployment.Annotations["flipper.example.com/restartedByCR"] != ""
			restartedByCRDKind := deployment.Annotations["flipper.example.com/restartedByCRDKind"] == "rollingupdate"

			return restartedAtExists && restartedBy && restartedByCRExists && restartedByCRDKind
		}, 2*time.Minute, 5*time.Second).Should(BeTrue())

		// Construct a label selector that matches all labels in the deployment's pod template
		selector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector()

		// List all pods in the namespace that match the label selector
		podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		Expect(err).NotTo(HaveOccurred())

		for _, pod := range podList.Items {
			// Verify the pod's start time
			Expect(pod.Status.StartTime.Time).To(BeTemporally(">", time.Now().Add(-5*time.Minute), time.Second))

			// Verify annotations related to restart information
			Eventually(func() bool {
				// Check if the annotation exists and is not empty
				restartedAtExists := pod.Annotations["kubectl.kubernetes.io/restartedAt"] != ""
				restartedBy := pod.Annotations["kubectl.kubernetes.io/restartedBy"] == "flipper-operator"
				restartedByCRExists := pod.Annotations["flipper.example.com/restartedByCR"] != ""
				restartedByCRDKind := pod.Annotations["flipper.example.com/restartedByCRDKind"] == "rollingupdate"

				return restartedAtExists && restartedBy && restartedByCRExists && restartedByCRDKind
			}, 2*time.Minute, 5*time.Second).Should(BeTrue())
		}
	}

	// Helper function to verify RollingUpdate CR and deployment pods
	verifyRollingUpdate := func(name, namespace string, interval time.Duration, expectedDeployments []string) {
		// Define GVR for RollingUpdate
		rollingUpdateGVR := schema.GroupVersionResource{
			Group:    "flipper.example.com",
			Version:  "v1alpha1",
			Resource: "rollingupdates",
		}

		var rollingUpdate *unstructured.Unstructured
		Eventually(func() error {
			var err error
			rollingUpdate, err = dynamicClient.Resource(rollingUpdateGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			_, found, err := unstructured.NestedFieldCopy(rollingUpdate.Object, "status", "lastRolloutTime")
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("lastRolloutTime not found")
			}

			deploymentsList, found, err := unstructured.NestedStringSlice(rollingUpdate.Object, "status", "deployments")
			if err != nil {
				return err
			}
			if len(expectedDeployments) == 0 {
				if found {
					return fmt.Errorf("deployments found")
				}
			} else {
				if !found {
					return fmt.Errorf("deployments not found")
				}
			}

			// equalStringSlicesWithoutOrder checks if two slices of strings are equal (same elements, regardless of order)
			equalStringSlicesWithoutOrder := func(a, b []string) bool {
				if len(a) != len(b) {
					return false
				}

				// Sort both slices
				sortedA := make([]string, len(a))
				sortedB := make([]string, len(b))
				copy(sortedA, a)
				copy(sortedB, b)
				sort.Strings(sortedA)
				sort.Strings(sortedB)

				// Compare sorted slices
				for i := range sortedA {
					if sortedA[i] != sortedB[i] {
						return false
					}
				}

				return true
			}

			if !equalStringSlicesWithoutOrder(deploymentsList, expectedDeployments) {
				return fmt.Errorf("deploymentsList does not match expectedDeployments")
			}

			return nil
		}, 5*time.Minute, 5*time.Second).Should(Succeed())

		// Check if the deployments' pods have restarted
		for _, deploymentName := range expectedDeployments {
			checkDeploymentPodsRestarted(deploymentName, namespace)
		}
	}

	checkDeploymentPodsReady := func(deploymentNames []string, namespace string) bool {
		for _, deploymentName := range deploymentNames {
			deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
			if err != nil {
				return false
			}

			// Construct a label selector that matches all labels in the deployment's pod template
			selector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector()

			// List all pods in the namespace that match the label selector
			podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: selector.String(),
			})
			if err != nil {
				return false
			}

			// Check readiness for each pod in the deployment
			for _, pod := range podList.Items {
				for _, condition := range pod.Status.Conditions {
					if condition.Type == corev1.PodReady && condition.Status != corev1.ConditionTrue {
						return false
					}
				}
			}
		}

		return true
	}

	Context("Namespace test-1", func() {
		It("should validate RollingUpdate CR and deployment pods test-1 namespace", func() {
			namespace := "test-1"
			expectedDeployments := []string{"nginx-deployment-v1", "nginx-deployment-v2", "nginx-deployment-v3"}

			// Apply the deployment YAML
			By("applying the deployment YAML")
			applyManifest("config/test/deployment-1.yaml")

			// Wait for deployment pods to be running
			By("waiting for deployment pods to be running")
			Eventually(func() bool {
				return checkDeploymentPodsReady(expectedDeployments, namespace)
			}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Deployment pods did not become ready in time")

			// Apply the RollingUpdate CR YAML
			By("applying the RollingUpdate CR YAML")
			applyManifest("config/test/rollingupdate-1.yaml")

			time.Sleep(30)
			// Wait for deployment pods to be rolling restart
			By("waiting for deployment pods to be running")
			Eventually(func() bool {
				return checkDeploymentPodsReady(expectedDeployments, namespace)
			}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Deployment pods did not become ready in time")

			verifyRollingUpdate("rollingupdate-1", namespace, 3*time.Minute, expectedDeployments)
		})
	})

	Context("Namespace test-2", func() {
		It("should validate RollingUpdate CR and deployment pods test-2 namespace", func() {
			namespace := "test-2"
			expectedDeployments := []string{"nginx-deployment"}

			// Apply the deployment YAML
			By("applying the deployment YAML")
			applyManifest("config/test/deployment-2.yaml")

			// Wait for deployment pods to be running
			By("waiting for deployment pods to be running")
			Eventually(func() bool {
				return checkDeploymentPodsReady(expectedDeployments, namespace)
			}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Deployment pods did not become ready in time")

			// Apply the RollingUpdate CR YAML
			By("applying the RollingUpdate CR YAML")
			applyManifest("config/test/rollingupdate-2.yaml")

			time.Sleep(30)
			// Wait for deployment pods to be rolling restart
			By("waiting for deployment pods to be running")
			Eventually(func() bool {
				return checkDeploymentPodsReady(expectedDeployments, namespace)
			}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Deployment pods did not become ready in time")

			verifyRollingUpdate("rollingupdate-2", namespace, 5*time.Minute, expectedDeployments)
		})
	})

	Context("Namespace test-3", func() {
		It("should validate RollingUpdate CR and deployment pods", func() {
			namespace := "test-3"
			expectedDeployments := []string{}

			// Apply the deployment YAML
			By("applying the deployment YAML")
			applyManifest("config/test/deployment-3.yaml")

			// Wait for deployment pods to be running
			By("waiting for deployment pods to be running")
			Eventually(func() bool {
				return checkDeploymentPodsReady(expectedDeployments, namespace)
			}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Deployment pods did not become ready in time")

			// Apply the RollingUpdate CR YAML
			By("applying the RollingUpdate CR YAML")
			applyManifest("config/test/rollingupdate-3.yaml")

			time.Sleep(30)
			// Wait for deployment pods to be rolling restart
			By("waiting for deployment pods to be running")
			Eventually(func() bool {
				return checkDeploymentPodsReady(expectedDeployments, namespace)
			}, 5*time.Minute, 5*time.Second).Should(BeTrue(), "Deployment pods did not become ready in time")

			verifyRollingUpdate("rollingupdate-3", namespace, 10*time.Minute, expectedDeployments)
		})
	})
})
