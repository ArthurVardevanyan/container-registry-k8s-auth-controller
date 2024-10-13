/*
Copyright 2024.

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
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1beta1 "github.com/ArthurVardevanyan/container-registry-k8s-auth-controller/api/v1beta1"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("Container Registry", func() {

	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var Audiences = []string{"openshift"}

	var ObjectName = getEnv("OBJECT_NAME", "test")
	var ObjectNamespace = getEnv("OBJECT_NAMESPACE", "smoke-tests")

	var ConfigName = getEnv("CONFIG_NAME", "google-wif-config")
	var RegistryLocation = getEnv("REGISTRY_LOCATION", "us-central1")
	var SecretName = getEnv("SECRET_NAME", "container-registry-auth-test")
	var ServiceAccount = getEnv("SERVICE_ACCOUNT", "wif-test")

	var GoogleServiceAccount = getEnv("GOOGLE_SERVICE_ACCOUNT", "wif-test@afr-operator-5560235161.iam.gserviceaccount.com")
	var GooglePoolProject = getEnv("GOOGLE_POOL_PROJECT", "448527874743")
	var GooglePoolName = getEnv("GOOGLE_POOL_NAME", "afr-operator-pool")
	var GoogleProviderName = getEnv("GOOGLE_PROVIDER_NAME", "afr-operator-provider")

	var RobotAccount = getEnv("ROBOT_ACCOUNT", "arthurvardevanyan+push")

	Context("Creating an Auth Object For Quay", func() {
		It("Should Read the Quay Wif Configs, and Create a Secret with a Short Lived Token", func() {
			By("By creating a new Container Registry Auth Object")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "quay",
					Quay: containerregistryv1beta1.Quay{
						RobotAccount: RobotAccount,
						URL:          "quay.io",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}
			createdObject := &containerregistryv1beta1.Auth{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Spec.SecretName).Should(Equal(SecretName))

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookUpKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.

			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)

		})

		It("Should Read the Quay Wif Configs, and and Failed on a missing service account", func() {
			By("By creating a new Container Registry Auth Object")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    "dne",
					Audiences:         Audiences,
					ContainerRegistry: "quay",
					Quay: containerregistryv1beta1.Quay{
						RobotAccount: RobotAccount,
						URL:          "quay.io",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}
			createdObject := &containerregistryv1beta1.Auth{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Spec.SecretName).Should(Equal(SecretName))

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return createdObject.Status.Error != ""
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Status.Error).Should(Equal("service Account 'dne' Not Found. Error: ServiceAccount \"dne\" not found"))
			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)
		})

		It("Should Read the Quay Wif Configs, and and Failed on a missing service account", func() {
			By("By creating a new Container Registry Auth Object")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    "pipeline",
					Audiences:         Audiences,
					ContainerRegistry: "quay",
					Quay: containerregistryv1beta1.Quay{
						RobotAccount: RobotAccount,
						URL:          "quay.io",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}
			createdObject := &containerregistryv1beta1.Auth{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Spec.SecretName).Should(Equal(SecretName))

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return createdObject.Status.Error != ""
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Status.Error).Should(Equal("Unable to Generate Quay Token: 400 Bad Request"))
			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)
		})

	})

	Context("Creating an Auth Object with Credential File", func() {
		It("Should Read a WIF ConfigMap, and Create a Secret with a Short Lived Token", func() {
			By("By creating a new Artifact Registry Auth Object")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "googleArtifactRegistry",
					GoogleArtifactRegistry: containerregistryv1beta1.GoogleArtifactRegistry{
						RegistryLocation: RegistryLocation,
						FileName:         "credentials_config.json",
						ObjectName:       ConfigName,
						Type:             "configMap",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}
			createdObject := &containerregistryv1beta1.Auth{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Spec.SecretName).Should(Equal(SecretName))

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookUpKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.

			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)

		})

		It("Should Return an Error for Invalid ConfigMap file", func() {
			By("By creating an Artifact Registry Auth Object with an Invalid ConfigMap Fle")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "googleArtifactRegistry",
					GoogleArtifactRegistry: containerregistryv1beta1.GoogleArtifactRegistry{
						RegistryLocation: RegistryLocation,
						FileName:         "credentials_config-bad.json",
						ObjectName:       ConfigName,
						Type:             "configMap",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			createdObject := &containerregistryv1beta1.Auth{}
			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return createdObject.Status.Error != ""
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Status.Error).Should(Equal("configMap key 'credentials_config-bad.json' not found. Error: <nil>"))
			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)
		})
	})

	Context("Creating an Auth Object with Inline File", func() {
		It("Should Read a WIF ConfigMap, and Create a Secret with a Short Lived Token", func() {
			By("By creating a new Artifact Registry Auth Object")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "googleArtifactRegistry",
					GoogleArtifactRegistry: containerregistryv1beta1.GoogleArtifactRegistry{
						RegistryLocation:     RegistryLocation,
						GoogleServiceAccount: GoogleServiceAccount,
						GooglePoolProject:    GooglePoolProject,
						GooglePoolName:       GooglePoolName,
						GoogleProviderName:   GoogleProviderName,
						Type:                 "inline",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}
			createdObject := &containerregistryv1beta1.Auth{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Spec.SecretName).Should(Equal(SecretName))

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookUpKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.

			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)

		})

		It("Should Return an Error for Invalid ConfigMap file", func() {
			By("By creating an Artifact Registry Auth Object with an Invalid ConfigMap Fle")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "googleArtifactRegistry",
					GoogleArtifactRegistry: containerregistryv1beta1.GoogleArtifactRegistry{
						RegistryLocation: RegistryLocation,
						FileName:         "credentials_config-bad.json",
						ObjectName:       ConfigName,
						Type:             "configMap",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			createdObject := &containerregistryv1beta1.Auth{}
			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return createdObject.Status.Error != ""
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Status.Error).Should(Equal("configMap key 'credentials_config-bad.json' not found. Error: <nil>"))
			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)
		})
	})

	Context("Creating an Auth Object with Inline File", func() {
		It("Should Read a WIF ConfigMap, and Create a Secret with a Short Lived Token", func() {
			By("By creating a new Artifact Registry Auth Object")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "googleArtifactRegistry",
					GoogleArtifactRegistry: containerregistryv1beta1.GoogleArtifactRegistry{
						RegistryLocation:     RegistryLocation,
						GoogleServiceAccount: GoogleServiceAccount,
						GooglePoolProject:    GooglePoolProject,
						GooglePoolName:       GooglePoolName,
						GoogleProviderName:   GoogleProviderName,
						Type:                 "inline",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}
			createdObject := &containerregistryv1beta1.Auth{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Spec.SecretName).Should(Equal(SecretName))

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookUpKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.

			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)

		})

		It("Should Return an Error for Invalid ConfigMap file", func() {
			By("By creating an Artifact Registry Auth Object with an Invalid ConfigMap Fle")
			// ctx := context.Background()
			Auth := &containerregistryv1beta1.Auth{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "containerregistry.arthurvardevanyan.com/v1beta1",
					Kind:       "Auth",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjectName,
					Namespace: ObjectNamespace,
				},
				Spec: containerregistryv1beta1.AuthSpec{
					SecretName:        SecretName,
					ServiceAccount:    ServiceAccount,
					Audiences:         Audiences,
					ContainerRegistry: "googleArtifactRegistry",
					GoogleArtifactRegistry: containerregistryv1beta1.GoogleArtifactRegistry{
						RegistryLocation: RegistryLocation,
						FileName:         "credentials_config-bad.json",
						ObjectName:       ConfigName,
						Type:             "configMap",
					},
				},
			}

			secretLookUpKey := types.NamespacedName{Name: SecretName, Namespace: ObjectNamespace}
			createdSecret := &v1.Secret{}

			k8sClient.Delete(ctx, Auth)
			k8sClient.Get(ctx, secretLookUpKey, createdSecret)
			k8sClient.Delete(ctx, createdSecret)

			Expect(k8sClient.Create(ctx, Auth)).Should(Succeed())

			createdObject := &containerregistryv1beta1.Auth{}
			objectLookUpKey := types.NamespacedName{Name: ObjectName, Namespace: ObjectNamespace}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				_ = k8sClient.Get(ctx, objectLookUpKey, createdObject)
				return createdObject.Status.Error != ""
			}, timeout, interval).Should(BeTrue())
			// Let's make sure our Schedule string value was properly converted/handled.
			Expect(createdObject.Status.Error).Should(Equal("configMap key 'credentials_config-bad.json' not found. Error: <nil>"))
			k8sClient.Delete(ctx, Auth)
			k8sClient.Delete(ctx, createdSecret)
		})
	})

})
