package kubernetes

import (
	"context"
	"fmt"

	authenticationV1 "k8s.io/api/authentication/v1"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Auth struct {
	client.Client
	Namespace               string
	FederatedServiceAccount string
	TokenExpirationSeconds  int
	Audiences               string
}

func New(
	client client.Client,
) Auth {
	return Auth{
		Client: client,
	}
}

func authToken(expirationSeconds int, audiences []string) *authenticationV1.TokenRequest {
	ExpirationSeconds := int64(expirationSeconds)

	tokenRequest := &authenticationV1.TokenRequest{
		Spec: authenticationV1.TokenRequestSpec{
			Audiences:         audiences,
			ExpirationSeconds: &ExpirationSeconds,
		},
	}

	return tokenRequest
}

func (r *Auth) GetKubernetesAuthToken(ctx context.Context, federatedServiceAccount string, namespace string, tokenExpirationSeconds int, audiences []string) (*authenticationV1.TokenRequest, error) {

	// Generate k8s Auth Token
	var serviceAccount coreV1.ServiceAccount
	k8sAuthToken := authToken(tokenExpirationSeconds, audiences)
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
