package agent

import (
	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	corev1 "k8s.io/api/core/v1"
)

//Add ISTIO_INIT_ENABLED env
func (a *Agent) createIstioInitEnv() []corev1.EnvVar {
	listEnvs := make([]corev1.EnvVar, 1)
	listEnvs[0].Name = "ISTIO_INIT_ENABLED"
	listEnvs[0].Value = "true"
	envs := a.istioEnvs(a.Annotations)
	for k, v := range envs {
		listEnvs = append(listEnvs, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	return listEnvs
}

//Add network_admin and network_raw to container
func (a *Agent) createIstioInitCapabilities() *corev1.Capabilities {
	cap := corev1.Capabilities{}
	cap.Add = append(cap.Add, "NET_ADMIN")
	cap.Add = append(cap.Add, "NET_RAW")
	return &cap
}

func (a *Agent) createIstioInitSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser:    pointerutil.Int64Ptr(0),
		RunAsGroup:   pointerutil.Int64Ptr(0),
		RunAsNonRoot: pointerutil.BoolPtr(false),
		Capabilities: a.createIstioInitCapabilities(),
	}
}

func (a *Agent) rewriteContainerCommand(cmd string) string {
	cmd += "&& bash /usr/local/bin/istio-init.sh"
	return cmd
}

func (a *Agent) CreateIstioInitSidecar() (corev1.Container, error) {

	arg := "bash /usr/local/bin/istio-init.sh"

	resources, err := a.parseResources()
	if err != nil {
		return corev1.Container{}, err
	}

	var pullPolicy corev1.PullPolicy
	pullPolicy = "Always"

	container := corev1.Container{
		Name:            "istio-agent-init",
		Image:           a.ImageName,
		Env:             a.createIstioInitEnv(),
		Resources:       resources,
		SecurityContext: a.createIstioInitSecurityContext(),
		Command:         []string{"/bin/sh", "-ec"},
		Args:            []string{arg},
		ImagePullPolicy: pullPolicy,
	}

	return container, nil
}
