package controller

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/pkg/errors"
	certsv1 "k8c.io/kubermatic-cert-manager-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *CertificateReconciler) certificateManager(ctx context.Context, config *certsv1.Certificate, secret *corev1.Secret) error {
	reqLogger := logf.FromContext(ctx, "Certificate Manager", "Track")
	reqLogger.Info("Reconciling Certificate Manager")

	err := r.Get(ctx, types.NamespacedName{Name: config.Spec.SecretRef.Name, Namespace: config.Namespace}, secret)
	if err == nil {
		reqLogger.Info("Certificate Manager secret already exists")
	} else if !apierrors.IsNotFound(err) {
		reqLogger.Error(err, "Failed to check if the Secret exists")
		return err
	}

	//secret does not exist, creating new certificate
	reqLogger.Info("Generating new Self Sign Certificate")

	// create self sign certificate steps
	privateKey, err := rsa.GenerateKey(rand.Reader, defaultKeySize)
	if err != nil {
		return errors.Wrap(err, "Failed to generate private key")
	}

	//create certificate template
	certificateTemplate, err := r.certificateTemplate(config)
	if err != nil {
		return errors.Wrap(err, "Failed to generate certificate template")
	}

	//create certificate
	certDer, err := x509.CreateCertificate(rand.Reader, certificateTemplate, certificateTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return errors.Wrap(err, "Failed to create certificate")
	}

	//Encode the certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// create the secret to store certificate
	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Spec.SecretRef.Name,
			Namespace: config.Namespace,
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       certPEM,
			corev1.TLSPrivateKeyKey: keyPEM,
		},
		Type: corev1.SecretTypeTLS,
	}

	//setting certificate as owner of secret
	if err = controllerutil.SetControllerReference(config, newSecret, r.Scheme); err != nil {
		reqLogger.Error(err, "Failed to set controller reference")
	}

	err = r.createOrUpdateSecret(ctx, config, newSecret)
	reqLogger.Info("Successfully created secret for certificate", "secret.Name", secret.Name)
	if err != nil {
		return errors.Wrap(err, "Failed to create or update secret")
	}

	return nil
}

func (r *CertificateReconciler) certificateTemplate(config *certsv1.Certificate) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to generate serial number")
	}

	notBefore := time.Now()
	validityDuration, err := parseCustomDuration(config.Spec.Validity)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse validity duration")
	}
	notAfter := notBefore.Add(validityDuration)

	certificateTemplate := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"k8c self-sign certificate"},
			CommonName:   config.Spec.DNSName,
		},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		DNSNames:              []string{config.Spec.DNSName},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	return certificateTemplate, nil
}

func (r *CertificateReconciler) createOrUpdateSecret(ctx context.Context, config *certsv1.Certificate, secret *corev1.Secret) error {
	err := r.Get(ctx, types.NamespacedName{Name: config.Spec.SecretRef.Name, Namespace: config.Namespace}, secret)
	if apierrors.IsNotFound(err) {
		err := r.Create(ctx, secret)
		if err != nil {
			return errors.Wrap(err, "creating secret object failed")
		}
	} else if err != nil {
		return errors.Wrap(err, "getting secret object failed")
	} else {
		// Secret exists, check if update and renewal is needed
		if !isSecretDataEqual(config, secret) {
			// renew and generate new certificate
			err = r.renewCertificate(ctx, config, secret)
			if err != nil {
				return errors.Wrap(err, "renewing secret object failed")
			}
		}
	}
	return nil
}

func (r *CertificateReconciler) renewCertificate(ctx context.Context, config *certsv1.Certificate, secret *corev1.Secret) error {
	// create self sign cerificate steps
	privateKey, err := rsa.GenerateKey(rand.Reader, defaultKeySize)
	if err != nil {
		return errors.Wrap(err, "Failed to generate private key")
	}

	//create certificate template
	certificateTemplate, err := r.certificateTemplate(config)
	if err != nil {
		return errors.Wrap(err, "Failed to generate certificate template")
	}

	//create certificate
	certDer, err := x509.CreateCertificate(rand.Reader, certificateTemplate, certificateTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return errors.Wrap(err, "Failed to create certificate")
	}

	//Ecnode the certificate
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	secret.Data[corev1.TLSCertKey] = certPEM
	secret.Data[corev1.TLSPrivateKeyKey] = keyPEM

	if err := r.Update(ctx, secret); err != nil {
		return errors.Wrap(err, "Failed to update secret")
	}
	return nil
}

func (r *CertificateReconciler) setConditionStatus(config *certsv1.Certificate, status metav1.ConditionStatus, conditionType, reason, message string) {
	conditionDetails := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             reason,
		Message:            message,
		ObservedGeneration: config.ObjectMeta.Generation,
	}

	if len(config.Status.Conditions) == 0 {
		config.Status.Conditions = []metav1.Condition{}
	}

	meta.SetStatusCondition(&config.Status.Conditions, conditionDetails)
}

func (r *CertificateReconciler) setCertificateStatus(ctx context.Context, config *certsv1.Certificate, secret *corev1.Secret) error {
	err := r.Get(ctx, types.NamespacedName{Name: config.Spec.SecretRef.Name, Namespace: config.Namespace}, secret)
	if err != nil {
		return errors.Wrap(err, "Failed to get existing secret")
	}

	certPEM := secret.Data[corev1.TLSCertKey]
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return errors.New("failed to decode certificate")
	}

	X509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse certificate")
	}

	config.Status.SerialNumber = X509Cert.SerialNumber.String()
	config.Status.Issuer = X509Cert.Issuer.CommonName
	config.Status.NotBefore = metav1.NewTime(X509Cert.NotBefore)
	config.Status.NotAfter = metav1.NewTime(X509Cert.NotAfter)

	r.setConditionStatus(config, metav1.ConditionTrue, TypeCertificateIssued, Success, CertificateIssued)
	r.setConditionStatus(config, metav1.ConditionTrue, TypeReconcileSuccess, ReconcileCompleted, ReconcileCompletedMessage)

	updatedErr := r.Status().Update(ctx, config)
	if updatedErr != nil {
		return errors.Wrap(updatedErr, "Failed to update status")
	}
	return nil
}
