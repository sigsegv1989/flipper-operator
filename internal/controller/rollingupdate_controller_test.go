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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flipperv1alpha1 "github.com/sigsegv1989/flipper-operator/api/v1alpha1"
)

var _ = Describe("RollingUpdate Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		rollingupdate := &flipperv1alpha1.RollingUpdate{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind RollingUpdate")
			err := k8sClient.Get(ctx, typeNamespacedName, rollingupdate)
			if err != nil && errors.IsNotFound(err) {
				resource := &flipperv1alpha1.RollingUpdate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: flipperv1alpha1.RollingUpdateSpec{
						MatchLabels: map[string]string{
							"key1": "value1",
						},
						Interval: "1m",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &flipperv1alpha1.RollingUpdate{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance RollingUpdate")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RollingUpdateReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(BeNumerically("==", time.Minute))

			rollingupdate := &flipperv1alpha1.RollingUpdate{}
			By("getting the custom resource for the Kind RollingUpdate")
			err = k8sClient.Get(ctx, typeNamespacedName, rollingupdate)
			Expect(err).NotTo(HaveOccurred())

			Expect(rollingupdate.Status.LastRolloutTime.IsZero()).NotTo(BeTrue())
			Expect(rollingupdate.Status.Deployments).To(BeEmpty())

			// Allow for some time discrepancy in the comparison
			lastRolloutTime := rollingupdate.Status.LastRolloutTime.Time
			now := time.Now()
			Expect(now.Sub(lastRolloutTime)).To(BeNumerically("<", time.Second*5))
		})
	})
})
