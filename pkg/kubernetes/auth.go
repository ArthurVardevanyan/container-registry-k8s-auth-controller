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
	Audience                string
}

func New(
	client client.Client,
	// namespace string,
	// federatedServiceAccount string,
	// tokenExpirationSeconds int,
	// audience string,
) Auth {
	return Auth{
		Client: client,
		// Namespace:               namespace,
		// FederatedServiceAccount: federatedServiceAccount,
		// TokenExpirationSeconds:  tokenExpirationSeconds,
		// Audience:                audience,
	}
}

func authToken(expirationSeconds int, audience string) *authenticationV1.TokenRequest {
	ExpirationSeconds := int64(expirationSeconds)

	tokenRequest := &authenticationV1.TokenRequest{
		Spec: authenticationV1.TokenRequestSpec{
			Audiences:         []string{audience},
			ExpirationSeconds: &ExpirationSeconds,
		},
	}

	return tokenRequest
}

func (r *Auth) GetKubernetesAuthToken(ctx context.Context, federatedServiceAccount string, namespace string, tokenExpirationSeconds int, audience string) (*authenticationV1.TokenRequest, error) {

	// Generate k8s Auth Token
	var serviceAccount coreV1.ServiceAccount
	k8sAuthToken := authToken(tokenExpirationSeconds, audience)
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
