/*
Copyright 2019 The OpenEBS Authors.

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

package webhook

import (
	"fmt"
	"os"
	"strings"

	menv "github.com/openebs/maya/pkg/env/v1alpha1"
	deploy "github.com/openebs/maya/pkg/kubernetes/deployment/appsv1/v1alpha1"
	secret "github.com/openebs/maya/pkg/kubernetes/secret"
	svc "github.com/openebs/maya/pkg/kubernetes/service/v1alpha1"
	validate "github.com/openebs/maya/pkg/kubernetes/webhook/validate/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	validatorServiceName = "admission-server-svc"
	validatorWebhook     = "openebs-validation-webhook-cfg"
	validatorSecret      = "admission-server-secret"
	webhookHandlerName   = "admission-webhook.openebs.io"
	validationPath       = "/validate"
	validationPort       = 8443
	// AdmissionNameEnvVar is the constant for env variable ADMISSION_WEBHOOK_NAME
	// which is the name of the current admission webhook
	AdmissionNameEnvVar = "ADMISSION_WEBHOOK_NAME"
	appCrt              = "app.crt"
	appKey              = "app.pem"
	rootCrt             = "ca.crt"
)

var (
	// TimeoutSeconds specifies the timeout for this webhook. After the timeout passes,
	// the webhook call will be ignored or the API call will fail based on the
	// failure policy.
	// The timeout value must be between 1 and 30 seconds.
	five = int32(5)
	// Ignore means that an error calling the webhook is ignored.
	Ignore = v1beta1.Ignore
)

// createWebhookService creates our webhook Service resource if it does not
// exist.
func createWebhookService(
	ownerReference metav1.OwnerReference,
	serviceName string,
	namespace string,
) error {

	_, err := svc.NewKubeClient(svc.WithNamespace(namespace)).
		Get(serviceName, metav1.GetOptions{})

	if err == nil {
		return nil
	}

	// error other than 'not found', return err
	if !k8serror.IsNotFound(err) {
		return errors.Wrapf(
			err,
			"failed to get webhook service {%v}",
			serviceName,
		)
	}

	// create service resource that refers to admission server pod
	serviceLabels := map[string]string{"app": "admission-webhook"}
	svcObj := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      serviceName,
			Labels: map[string]string{
				"app":                       "admission-webhook",
				"openebs.io/component-name": "admission-webhook-svc",
			},
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Spec: corev1.ServiceSpec{
			Selector: serviceLabels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       443,
					TargetPort: intstr.FromInt(validationPort),
				},
			},
		},
	}
	_, err = svc.NewKubeClient(svc.WithNamespace("openebs")).
		Create(svcObj)
	return err
}

// createAdmissionService creates our ValidatingWebhookConfiguration resource
// if it does not exist.
func createAdmissionService(
	ownerReference metav1.OwnerReference,
	validatorWebhook string,
	namespace string,
	serviceName string,
	signingCert []byte,
) error {

	_, err := GetValidatorWebhook(validatorWebhook)
	// validator object already present, no need to do anything
	if err == nil {
		return nil
	}

	// error other than 'not found', return err
	if !k8serror.IsNotFound(err) {
		return errors.Wrapf(
			err,
			"failed to get webhook validator {%v}",
			validatorWebhook,
		)
	}

	webhookHandler := v1beta1.Webhook{
		Name: webhookHandlerName,
		Rules: []v1beta1.RuleWithOperations{{
			Operations: []v1beta1.OperationType{
				v1beta1.Create,
				v1beta1.Delete,
			},
			Rule: v1beta1.Rule{
				APIGroups:   []string{"*"},
				APIVersions: []string{"*"},
				Resources:   []string{"persistentvolumeclaims"},
			},
		},
			{
				Operations: []v1beta1.OperationType{
					v1beta1.Create,
					v1beta1.Update,
				},
				Rule: v1beta1.Rule{
					APIGroups:   []string{"*"},
					APIVersions: []string{"*"},
					Resources:   []string{"cstorpoolclusters"},
				},
			}},
		ClientConfig: v1beta1.WebhookClientConfig{
			Service: &v1beta1.ServiceReference{
				Namespace: namespace,
				Name:      serviceName,
				Path:      StrPtr(validationPath),
			},
			CABundle: signingCert,
		},
		TimeoutSeconds: &five,
		FailurePolicy:  &Ignore,
	}

	validator := &v1beta1.ValidatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "validatingWebhookConfiguration",
			APIVersion: "admissionregistration.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: validatorWebhook,
			Labels: map[string]string{
				"app":                       "admission-webhook",
				"openebs.io/component-name": "admission-webhook",
			},
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Webhooks: []v1beta1.Webhook{webhookHandler},
	}

	_, err = validate.KubeClient().Create(validator)

	return err
}

// createCertsSecret creates a self-signed certificate and stores it as a
// secret resource in Kubernetes.
func createCertsSecret(
	ownerReference metav1.OwnerReference,
	secretName string,
	serviceName string,
	namespace string,
) (*corev1.Secret, error) {

	// Create a signing certificate
	caKeyPair, err := NewCA(fmt.Sprintf("%s-ca", serviceName))
	if err != nil {
		return nil, fmt.Errorf("failed to create root-ca: %v", err)
	}

	// Create app certs signed through the certificate created above
	apiServerKeyPair, err := NewServerKeyPair(
		caKeyPair,
		strings.Join([]string{serviceName, namespace, "svc"}, "."),
		serviceName,
		namespace,
		"cluster.local",
		[]string{},
		[]string{},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create server key pair: %v", err)
	}

	// create an opaque secret resource with certificate(s) created above
	secretObj := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                       "admission-webhook",
				"openebs.io/component-name": "admission-webhook",
			},
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			appCrt:  EncodeCertPEM(apiServerKeyPair.Cert),
			appKey:  EncodePrivateKeyPEM(apiServerKeyPair.Key),
			rootCrt: EncodeCertPEM(caKeyPair.Cert),
		},
	}

	return secret.NewKubeClient(secret.WithNamespace(namespace)).Create(secretObj)
}

// GetValidatorWebhook fetches the webhook validator resource in
// Openebs namespace.
func GetValidatorWebhook(
	validator string,
) (*v1beta1.ValidatingWebhookConfiguration, error) {

	return validate.KubeClient().Get(validator, metav1.GetOptions{})
}

// StrPtr convert a string to a pointer
func StrPtr(s string) *string {
	return &s
}

// InitValidationServer creates secret, service and admission validation k8s
// resources. All these resources are created in the same namespace where
// openebs components is running.
func InitValidationServer(
	ownerReference metav1.OwnerReference,
) error {

	// Fetch our namespace
	openebsNamespace, err := getOpenebsNamespace()
	if err != nil {
		return err
	}

	// Check to see if webhook secret is already present
	certSecret, err := GetSecret(openebsNamespace, validatorSecret)
	if err != nil {
		if k8serror.IsNotFound(err) {
			// Secret not found, create certs and the secret object
			certSecret, err = createCertsSecret(
				ownerReference,
				validatorSecret,
				validatorServiceName,
				openebsNamespace,
			)
			if err != nil {
				return fmt.Errorf(
					"failed to create secret(%s) resource %v",
					validatorSecret,
					err,
				)
			}
		} else {
			// Unable to read secret object
			return fmt.Errorf(
				"unable to read secret object %s: %v",
				validatorSecret,
				err,
			)
		}
	}

	signingCertBytes, ok := certSecret.Data[rootCrt]
	if !ok {
		return fmt.Errorf(
			"%s value not found in %s secret",
			rootCrt,
			validatorSecret,
		)
	}

	serviceErr := createWebhookService(
		ownerReference,
		validatorServiceName,
		openebsNamespace,
	)
	if serviceErr != nil {
		return fmt.Errorf(
			"failed to create Service{%s}: %v",
			validatorServiceName,
			serviceErr,
		)
	}

	validatorErr := createAdmissionService(
		ownerReference,
		validatorWebhook,
		openebsNamespace,
		validatorServiceName,
		signingCertBytes,
	)
	if validatorErr != nil {
		return fmt.Errorf(
			"failed to create validator{%s}: %v",
			validatorWebhook,
			validatorErr,
		)
	}

	return nil
}

// GetSecret fetches the secret resource in the given namespace.
func GetSecret(
	namespace string,
	secretName string,
) (*corev1.Secret, error) {

	return secret.NewKubeClient(secret.WithNamespace(namespace)).Get(secretName, metav1.GetOptions{})
}

// getOpenebsNamespace gets the namespace OPENEBS_NAMESPACE env value which is
// set by the downward API where admission server has been deployed
func getOpenebsNamespace() (string, error) {

	ns, found := menv.Lookup(menv.OpenEBSNamespace)
	if !found {
		return "", fmt.Errorf("%s must be set", menv.OpenEBSNamespace)
	}
	return ns, nil
}

// GetAdmissionName return the admission server name
func GetAdmissionName() (string, error) {
	admissionName, found := os.LookupEnv(AdmissionNameEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", AdmissionNameEnvVar)
	}
	if len(admissionName) == 0 {
		return "", fmt.Errorf("%s must not be empty", AdmissionNameEnvVar)
	}
	return admissionName, nil
}

// GetAdmissionReference is a utility function to fetch a reference
// to the admission webhook deployment object
func GetAdmissionReference() (*metav1.OwnerReference, error) {

	// Fetch our namespace
	openebsNamespace, err := getOpenebsNamespace()
	if err != nil {
		return nil, err
	}

	// Fetch our admission server deployment object
	admdeployList, err := deploy.NewKubeClient(deploy.WithNamespace(openebsNamespace)).
		List(&metav1.ListOptions{LabelSelector: webhookLabel})
	if err != nil {
		return nil, fmt.Errorf("failed to list admission deployment: %s", err.Error())
	}

	for _, admdeploy := range admdeployList.Items {
		if len(admdeploy.Name) != 0 {
			return metav1.NewControllerRef(admdeploy.GetObjectMeta(), schema.GroupVersionKind{
				Group:   appsv1.SchemeGroupVersion.Group,
				Version: appsv1.SchemeGroupVersion.Version,
				Kind:    "Deployment",
			}), nil

		}
	}
	return nil, fmt.Errorf("failed to create deployment ownerReference")
}
