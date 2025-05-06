/*
Copyright 2025.

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "github.com/bryantanderson/basic-kubernetes-operator/api/v1"
)

var _ = Describe("ConfigMapSync Controller", func() {
	Context("When reconciling a resource", func() {
		ctx := context.Background()

		resourceName := "testmap"

		sourceNamespaceName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "testone",
		}
		destinationNamespaceName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "testtwo",
		}

		configmapsync := &appsv1.ConfigMapSync{
			Spec: appsv1.ConfigMapSyncSpec{
				DestinationNamespace: destinationNamespaceName.Namespace,
				SourceNamespace:      sourceNamespaceName.Namespace,
				ConfigMapName:        resourceName,
			},
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ConfigMapSync")
			err := k8sClient.Get(ctx, sourceNamespaceName, configmapsync)
			if err != nil && errors.IsNotFound(err) {
				resource := &appsv1.ConfigMapSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      sourceNamespaceName.Name,
						Namespace: sourceNamespaceName.Namespace,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ConfigMapSyncReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: sourceNamespaceName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
