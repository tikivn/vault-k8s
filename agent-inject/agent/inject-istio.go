package agent


import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	meshconfig "istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pkg/config/mesh"
	"istio.io/istio/pkg/kube/inject"
	"istio.io/pkg/log"
	"istio.io/pkg/version"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)


const (
	configMapKey       = "mesh"
	injectConfigMapKey = "config"
	valuesConfigMapKey = "values"
	defaultIstioNamespace = "istio-system"
	defaultMeshConfigMapName   = "istio"
	defaultInjectConfigMapName = "istio-sidecar-injector"
)

var (
	//kubeconfig   = "/Users/uc.dang/.kube/dev"
	configContext = ""
	istioNamespace  	string
	meshConfigMapName   string
	valuesFile          string
	injectConfigMapName string
)

func initConfig(){
	meshConfigMapName = defaultMeshConfigMapName
	injectConfigMapName = defaultInjectConfigMapName
	if istioNamespace = os.Getenv("ISTIO_NAMESPACE") ; istioNamespace == "" {
		istioNamespace = defaultIstioNamespace
	}
}

func createInterface() (kubernetes.Interface, error) {
	initConfig()
	//restConfig, err := kube.BuildClientConfig(kubeconfig, configContext)
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Errorf("Error when read kubeconfig from service account: %s", err)
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

func getMeshConfigFromConfigMap(command string) (*meshconfig.MeshConfig, error) {
	client, err := createInterface()
	if err != nil {
		return nil, err
	}

	meshConfigMap, err := client.CoreV1().ConfigMaps(istioNamespace).Get(meshConfigMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not read valid configmap %q from namespace %q: %v - "+
			"Use --meshConfigFile or re-run "+command+" with `-i <istioSystemNamespace> and ensure valid MeshConfig exists",
			meshConfigMapName, istioNamespace, err)
	}
	// values in the data are strings, while proto might use a
	// different data type.  therefore, we have to get a value by a
	// key
	configYaml, exists := meshConfigMap.Data[configMapKey]
	if !exists {
		return nil, fmt.Errorf("missing configuration map key %q", configMapKey)
	}
	cfg, err := mesh.ApplyMeshConfigDefaults(configYaml)
	if err != nil {
		err = multierror.Append(err, fmt.Errorf("istioctl version %s cannot parse mesh config.  Install istioctl from the latest Istio release",
			version.Info.Version))
	}
	return cfg, err
}

func getInjectConfigFromConfigMap() (string, error) {
	client, err := createInterface()
	if err != nil {
		return "", err
	}

	meshConfigMap, err := client.CoreV1().ConfigMaps(istioNamespace).Get(injectConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("could not find valid configmap %q from namespace  %q: %v - "+
			"Use --injectConfigFile or re-run kube-inject with `-i <istioSystemNamespace> and ensure istio-inject configmap exists",
			injectConfigMapName, istioNamespace, err)
	}
	// values in the data are strings, while proto might use a
	// different data type.  therefore, we have to get a value by a
	// key
	injectData, exists := meshConfigMap.Data[injectConfigMapKey]
	if !exists {
		return "", fmt.Errorf("missing configuration map key %q in %q",
			injectConfigMapKey, injectConfigMapName)
	}
	var injectConfig inject.Config
	if err := yaml.Unmarshal([]byte(injectData), &injectConfig); err != nil {
		return "", fmt.Errorf("unable to convert data from configmap %q: %v",
			injectConfigMapName, err)
	}
	log.Debugf("using inject template from configmap %q", injectConfigMapName)
	return injectConfig.Template, nil
}
// grabs the raw values from the ConfigMap. These are encoded as JSON.
func getValuesFromConfigMap() (string, error) {
	client, err := createInterface()
	if err != nil {
		return "", err
	}

	meshConfigMap, err := client.CoreV1().ConfigMaps(istioNamespace).Get(injectConfigMapName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("could not find valid configmap %q from namespace  %q: %v - "+
			"Use --injectConfigFile or re-run kube-inject with `-i <istioSystemNamespace> and ensure istio-inject configmap exists",
			injectConfigMapName, istioNamespace, err)
	}

	valuesData, exists := meshConfigMap.Data[valuesConfigMapKey]
	if !exists {
		return "", fmt.Errorf("missing configuration map key %q in %q",
			valuesConfigMapKey, injectConfigMapName)
	}

	return valuesData, nil
}

func GetInjectContainer(pod *corev1.Pod) (*inject.SidecarInjectionSpec ,error) {

	var meshConfig *meshconfig.MeshConfig
	var sidecarTemplate string
	var valuesConfig string
	var err error
	if meshConfig, err = getMeshConfigFromConfigMap("kube-inject"); err != nil {
		return nil, err
	}
	if sidecarTemplate, err = getInjectConfigFromConfigMap(); err != nil {
		return nil, err
	}
	if valuesConfig, err = getValuesFromConfigMap(); err != nil {
		return nil, err
	}

	var deploymentMetadata *metav1.ObjectMeta
	var metadata *metav1.ObjectMeta
	var podSpec *corev1.PodSpec
	var typeMeta *metav1.TypeMeta


	typeMeta = &pod.TypeMeta
	metadata = &pod.ObjectMeta
	deploymentMetadata = &pod.ObjectMeta
	podSpec = &pod.Spec

	spec, _, err := inject.InjectionData(
		sidecarTemplate,
		valuesConfig,
		SidecarTemplateVersionHash(sidecarTemplate),
		typeMeta,
		deploymentMetadata,
		podSpec,
		metadata,
		meshConfig.DefaultConfig,
		meshConfig)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func SidecarTemplateVersionHash(in string) string {
	hash := sha256.Sum256([]byte(in))
	return hex.EncodeToString(hash[:])
}