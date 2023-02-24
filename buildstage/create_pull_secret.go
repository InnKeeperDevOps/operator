package buildstage

import (
	b64 "encoding/base64"
	cicdv1alpha1 "github.com/innkeeperdevops/operator/api/v1alpha1"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *BuildStage) CreatePullSecret(deploy *cicdv1alpha1.BuildDeploy, username string, password string, namespace string) *v12.Secret {
	base64Auths := b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	dockerAuth := "{\"auths\":{\"" + deploy.Spec.Publish.Host + "\":{\"username\":\"" + username + "\",\"password\":\"" + password + "\",\"auth\":\"" + base64Auths + "\"}}}"
	secret := &v12.Secret{
		Type: "kubernetes.io/dockerconfigjson",
		ObjectMeta: v1.ObjectMeta{
			Name:      "docker-pull-" + deploy.Spec.Publish.Host,
			Namespace: namespace,
		},
		Data: map[string][]byte{v12.DockerConfigJsonKey: []byte(dockerAuth)},
	}
	println(dockerAuth)
	return secret
}
