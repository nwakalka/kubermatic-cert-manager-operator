package controller

import (
	"crypto/x509"
	"encoding/pem"
	"strconv"
	"strings"
	"time"

	certsv1 "k8c.io/kubermatic-cert-manager-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	day            = "d"
	finalizerName  = "certs.k8c.io/finalizer"
	defaultKeySize = 2048

	TypeCertificateIssued = "Issued"
	TypeReconcileError    = "ReconciliationError"
	ReconcileFailed       = "ReconcileFailed"
	TypeReconcileSuccess  = "ReconciliationSuccess"
	ReconcileCompleted    = "ReconcileCompleted"

	Success                   = "Success"
	CertificateIssued         = "Successfully Issued Certificate"
	ReconcileCompletedMessage = "Reconcile Completed Successfully"
)

// Helper function to parse duration
func parseCustomDuration(duration string) (time.Duration, error) {
	//check if "d" suffix in duration
	if strings.HasSuffix(duration, day) {
		//convert to int
		daysStr := strings.TrimSuffix(duration, day)
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			//add error wrap
			return 0, err
		}
		//convert days to hours
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(duration)
}

// Helper function to compare secret data
func isSecretDataEqual(config *certsv1.Certificate, secret *corev1.Secret) bool {
	certPEM, ok := secret.Data[corev1.TLSCertKey]
	if !ok {
		return false
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	if x509Cert.Subject.CommonName != config.Spec.DNSName {
		return false
	}

	// Check if certificate is expiring within a certain threshold (e.g., 30 days)
	renewalThreshold := 30 * 24 * time.Hour
	return time.Until(x509Cert.NotAfter) >= renewalThreshold
}
