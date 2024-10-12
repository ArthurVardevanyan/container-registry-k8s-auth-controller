package kubernetes

import (
	b64 "encoding/base64"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	coreV1 "k8s.io/api/core/v1"
)

func ImagePullSecretConfig(userName string, token string, url string) string {
	const ImagePullSecretTemplate = "{\"auths\": {\"REGISTRY\": {\"auth\": \"BASE64TOKEN\"}}}"
	BASE64TOKEN := b64.StdEncoding.EncodeToString([]byte(userName + ":" + token))
	ImagePullSecret := strings.Replace(ImagePullSecretTemplate, "REGISTRY", url, 1)
	ImagePullSecret = strings.Replace(ImagePullSecret, "BASE64TOKEN", BASE64TOKEN, 1)

	return ImagePullSecret
}

func ImagePullSecretObject(name string, namespace string, dockerConfig string, ownerReference []metaV1.OwnerReference) *coreV1.Secret {
	// https://stackoverflow.com/questions/64758486/how-to-create-docker-secret-with-client-go
	secret := &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			OwnerReferences: ownerReference,
		},
		Type:       "kubernetes.io/dockerconfigjson",
		StringData: map[string]string{".dockerconfigjson": dockerConfig},
	}

	return secret
}
