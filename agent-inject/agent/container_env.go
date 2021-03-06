package agent

import (
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"
)

// ContainerEnvVars adds the applicable environment vars
// for the Vault Agent sidecar.
func (a *Agent) ContainerEnvVars(init bool, isExisted bool) ([]corev1.EnvVar, error) {
	var envs []corev1.EnvVar

	if a.Vault.ClientTimeout != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "VAULT_CLIENT_TIMEOUT",
			Value: a.Vault.ClientTimeout,
		})
	}

	if a.Vault.ClientMaxRetries != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "VAULT_MAX_RETRIES",
			Value: a.Vault.ClientMaxRetries,
		})
	}

	if a.Inject {
		envs = append(envs, corev1.EnvVar{
			Name:  "VAULT_ENABLED",
			Value: "true",
		})
	}

	if a.ConfigMapName == "" {
		config, err := a.newConfig(init)
		if err != nil {
			return envs, err
		}

		b64Config := base64.StdEncoding.EncodeToString(config)
		envs = append(envs, corev1.EnvVar{
			Name:  "VAULT_CONFIG",
			Value: b64Config,
		})
	}

	for _, envVar := range a.PlutonEnvs {
		envs = append(envs, corev1.EnvVar{
			Name:  envVar.Key,
			Value: envVar.Value,
		})
	}

	if isExisted {
		envs = append(envs, corev1.EnvVar{
			Name:  "TK_TDAGENT_ENABLED",
			Value: "false",
		})
		envs = append(envs, corev1.EnvVar{
			Name:  "TK_TELEGRAF_ENABLED",
			Value: "false",
		})
	}

	return envs, nil
}
