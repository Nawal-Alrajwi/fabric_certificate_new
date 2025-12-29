package chaincode

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3" // استخدام SHA-3 المتطورة
	"encoding/hex"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// الهيكل المطور للشهادة
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Major       string `json:"Major"`
	University  string `json:"University"`
	IssueDate   string `json:"IssueDate"`
	Grade       string `json:"Grade"`
	IssuerID    string `json:"IssuerID"`
	CertHash    string `json:"CertHash"` // حقل البصمة الرقمية الجديد
}

// دالة داخلية لتوليد بصمة SHA-3
func calculateSHA3Hash(data string) string {
	hash := sha3.New256()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

// IssueCertificate المطور
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, major string, university string, issueDate string, grade string, issuerID string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
	}

	// دمج البيانات لتوليد البصمة
	combinedData := fmt.Sprintf("%s%s%s%s", id, studentName, university, issueDate)
	certHash := calculateSHA3Hash(combinedData)

	cert := Certificate{
		ID:          id,
		StudentName: studentName,
		Major:       major,
		University:  university,
		IssueDate:   issueDate,
		Grade:       grade,
		IssuerID:    issuerID,
		CertHash:    certHash,
	}

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// VerifyCertificate المطور (لتقليل الـ Latency)
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, providedData string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("فشل في قراءة بيانات الحالة: %v", err)
	}
	if certJSON == nil {
		return false, fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return false, err
	}

	// التحقق السريع عبر مقارنة بصمة البيانات المدخلة مع البصمة المخزنة
	currentHash := calculateSHA3Hash(providedData)
	return cert.CertHash == currentHash, nil
}

// CertificateExists دالة مساعدة
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return certJSON != nil, nil
}
