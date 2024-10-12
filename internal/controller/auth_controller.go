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
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	b64 "encoding/base64"

	"encoding/json"

	authenticationV1 "k8s.io/api/authentication/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	containerregistryv1beta1 "github.com/ArthurVardevanyan/container-registry-k8s-auth-controller/api/v1beta1"
)

func BoolPointer(b bool) *bool {
	return &b
}

// AuthReconciler reconciles a Auth object
type AuthReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func imagePullSecretConfig(userName string, TOKEN string) string {
	const ImagePullSecretTemplate = "{\"auths\": {\"REGISTRY\": {\"auth\": \"BASE64TOKEN\"}}}"
	ImagePullSecret := strings.Replace(ImagePullSecretTemplate, "REGISTRY", "quay.io", 1)
	BASE64TOKEN := b64.StdEncoding.EncodeToString([]byte(userName + ":" + TOKEN))
	ImagePullSecret = strings.Replace(ImagePullSecret, "BASE64TOKEN", BASE64TOKEN, 1)

	return ImagePullSecret
}

func imagePullSecretObject(name string, namespace string, dockerConfig string) *coreV1.Secret {
	// https://stackoverflow.com/questions/64758486/how-to-create-docker-secret-with-client-go
	secret := &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type:       "kubernetes.io/dockerconfigjson",
		StringData: map[string]string{".dockerconfigjson": dockerConfig},
	}

	return secret
}

func getQuayRobotToken(fedToken string, robotAccount string) (string, error) {
	url := "http://quay.io/oauth2/federation/robot/token"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.SetBasicAuth(robotAccount, fedToken)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	// fmt.Println("Response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var result map[string]interface{}

	json.Unmarshal([]byte(body), &result)

	return result["token"].(string), nil
}

func updateArtifactRegistryObject(r *AuthReconciler, reconcilerContext context.Context, containerRegistryAuth containerregistryv1beta1.Auth, expirationSeconds int) (ctrl.Result, error) {
	if err := r.Status().Update(reconcilerContext, &containerRegistryAuth); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to update Artifact Registry Auth status: %w", err)
	} else {
		if expirationSeconds == 0 {
			expirationSeconds = 36000
		}
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(expirationSeconds-60)}, nil
	}
}

func kubernetesAuthToken(expirationSeconds int) *authenticationV1.TokenRequest {
	ExpirationSeconds := int64(expirationSeconds)

	tokenRequest := &authenticationV1.TokenRequest{
		Spec: authenticationV1.TokenRequestSpec{
			Audiences:         []string{"openshift"},
			ExpirationSeconds: &ExpirationSeconds,
		},
	}

	return tokenRequest
}

func (r *AuthReconciler) getKubernetesAuthToken(ctx context.Context, federatedServiceAccount string, namespace string, tokenExpirationSeconds int) (*authenticationV1.TokenRequest, error) {

	// Generate k8s Auth Token
	var serviceAccount coreV1.ServiceAccount
	k8sAuthToken := kubernetesAuthToken(tokenExpirationSeconds)
	err := r.Get(ctx, client.ObjectKey{Name: federatedServiceAccount, Namespace: namespace}, &serviceAccount)
	if err != nil {

		return nil, fmt.Errorf("service Account '%s' Not Found. Error: %v", federatedServiceAccount, err)
	}
	err = r.SubResource("token").Create(ctx, &serviceAccount, k8sAuthToken)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes token. Error: %v", err)
	}

	return k8sAuthToken, nil
}

// +kubebuilder:rbac:groups=containerregistry.arthurvardevanyan.com,resources=auths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=containerregistry.arthurvardevanyan.com,resources=auths/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=containerregistry.arthurvardevanyan.com,resources=auths/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Auth object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *AuthReconciler) Reconcile(reconcilerContext context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(reconcilerContext)
	log.V(1).Info(req.Name)

	// Common Variables
	const tokenExpirationSeconds = 3600
	var err error
	var error string

	// Incept Object
	var containerRegistryAuth containerregistryv1beta1.Auth
	if err = r.Get(reconcilerContext, req.NamespacedName, &containerRegistryAuth); err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.V(1).Info("Artifact Registry Auth Object Not Found or No Longer Exists!")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "Unable to fetch Artifact Registry Auth Object")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	//Reset Error
	containerRegistryAuth.Status.Error = ""

	kubernetesToken, err := r.getKubernetesAuthToken(reconcilerContext, containerRegistryAuth.Spec.ServiceAccount, req.NamespacedName.Namespace, tokenExpirationSeconds)
	quayToken, err := getQuayRobotToken(kubernetesToken.Status.Token, containerRegistryAuth.Spec.Quay.RobotAccount)

	// Create Image Pull Secret
	dockerConfig := imagePullSecretConfig(containerRegistryAuth.Spec.Quay.RobotAccount, quayToken)
	imagePullSecret := imagePullSecretObject(containerRegistryAuth.Spec.SecretName, req.NamespacedName.Namespace, dockerConfig)
	err = r.Update(reconcilerContext, imagePullSecret)
	if err != nil {
		err = r.Create(reconcilerContext, imagePullSecret)
		if err != nil {
			error = "Unable to Create Image Pull Secret"
			containerRegistryAuth.Status.Error = error
			log.Error(err, error)
			return updateArtifactRegistryObject(r, reconcilerContext, containerRegistryAuth, 0)
		}
	}

	return updateArtifactRegistryObject(r, reconcilerContext, containerRegistryAuth, tokenExpirationSeconds)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&containerregistryv1beta1.Auth{}).
		Complete(r)
}
