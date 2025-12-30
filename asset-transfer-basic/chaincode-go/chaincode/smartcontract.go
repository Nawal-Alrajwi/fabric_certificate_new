package chaincode

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"golang.org/x/crypto/sha3" // استخدام SHA-3 المتطورة لتوفير أمن أعلى من SHA-256 التقليدي
)

// SmartContract defines the structure for our certificate logic
type SmartContract struct {
	contractapi.Contract
}

// Certificate يمثل هيكل الشهادة المطور
// ملاحظة: تم تصحيح علامات الـ Backticks لضمان قبول الكود في Go
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Major       string `json:"Major"`
	University  string `json:"University"`
	IssueDate   string `json:"IssueDate"`
	Grade       string `json:"Grade"`
	IssuerID    string `json:"IssuerID"`
	CertHash    string `json:"CertHash"` // البصمة الرقمية المضافة لضمان سرعة التحقق
}

// PaginatedQueryResult يمثل هيكل الرد المقسم لعمليات الاستعلام الضخمة
type PaginatedQueryResult struct {
	Records             []*Certificate `json:"records"`
	FetchedRecordsCount int32          `json:"fetchedRecordsCount"`
	Bookmark            string         `json:"bookmark"`
}

// calculateSHA3Hash دالة داخلية لتوليد بصمة SHA-3
func calculateSHA3Hash(data string) string {
	hash := sha3.New256()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

// IssueCertificate: إصدار شهادة جديدة مع توليد بصمة SHA-3 وحفظها في السجل
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, major string, university string, issueDate string, grade string, issuerID string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
	}

	// دمج البيانات الأساسية لتوليد البصمة
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

// VerifyCertificate: التحقق الفوري من صحة الشهادة بمطابقة بصمة SHA-3
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string, providedData string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("فشل في قراءة البيانات: %v", err)
	}
	if certJSON == nil {
		return false, fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return false, err
	}

	// حساب بصمة البيانات المقدمة ومقارنتها بالمخزنة
	currentHash := calculateSHA3Hash(providedData)
	return cert.CertHash == currentHash, nil
}

// GetAllAssetsWithPagination: استعلام مطور يدعم التقسيم لحل مشكلة الـ Timeout عند 100 TPS
func (s *SmartContract) GetAllAssetsWithPagination(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedQueryResult, error) {
	resultsIterator, responseMetadata, err := ctx.GetStub().GetStateByRangeWithPagination("", "", pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var certs []*Certificate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var cert Certificate
		err = json.Unmarshal(queryResponse.Value, &cert)
		if err != nil {
			return nil, err
		}
		certs = append(certs, &cert)
	}

	return &PaginatedQueryResult{
		Records:             certs,
		FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
		Bookmark:            responseMetadata.Bookmark,
	}, nil
}

// DeleteAsset: إلغاء (حذف) الشهادة من السجل
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	return ctx.GetStub().DelState(id)
}

// CertificateExists: التحقق من وجود السجل في قاعدة البيانات
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return certJSON != nil, nil
}
