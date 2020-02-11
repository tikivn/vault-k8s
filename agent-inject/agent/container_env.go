package agent

import (
	"encoding/base64"

	corev1 "k8s.io/api/core/v1"
)

// ContainerEnvVars adds the applicable environment vars
// for the Vault Agent sidecar.
func (a *Agent) ContainerEnvVars(init bool) ([]corev1.EnvVar, error) {
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

	if a.Pluton.InfluxdbUrl != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "TK_INFLUXDB_URL",
			Value: a.Pluton.InfluxdbUrl,
		})
	}

	if a.Inject {
		envs = append(envs, corev1.EnvVar{
			Name:  "VAULT_ENABLED",
			Value: "true",
		})
	}

	if a.ConfigMapName == "" && !a.InjectPluton {
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

	return envs, nil
}
