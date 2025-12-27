package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// CertificateSmartContract لخدمة إدارة الشهادات الإلكترونية
type CertificateSmartContract struct {
	contractapi.Contract
}

// Certificate يمثل هيكل الشهادة الرقمية المخزنة على البلوكشين
type Certificate struct {
	ID             string `json:"ID"`             // الرقم التسلسلي للشهادة
	StudentName    string `json:"StudentName"`    // اسم الطالب
	Major          string `json:"Major"`          // التخصص العلمي
	University     string `json:"University"`     // الجامعة المصدرة
	IssueDate      string `json:"IssueDate"`      // تاريخ الإصدار
	Grade          string `json:"Grade"`          // التقدير العام
	IssuerID       string `json:"IssuerID"`       // معرف الجهة الموقعة (للحماية)
}

// IssueCertificate إصدار شهادة جديدة وإضافتها إلى الليدجر
func (s *CertificateSmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, major string, university string, issueDate string, grade string, issuerID string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
	}

	cert := Certificate{
		ID:          id,
		StudentName: studentName,
		Major:       major,
		University:  university,
		IssueDate:   issueDate,
		Grade:       grade,
		IssuerID:    issuerID,
	}
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// QueryCertificate استرجاع بيانات شهادة معينة باستخدام الرقم التسلسلي
func (s *CertificateSmartContract) QueryCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("فشل في القراءة من البلوكشين: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// RevokeCertificate إلغاء شهادة (حذفها من الحالة الحالية) في حالة التزوير أو الخطأ
func (s *CertificateSmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("الشهادة %s غير موجودة لإلغائها", id)
	}

	return ctx.GetStub().DelState(id)
}

// CertificateExists التحقق من وجود الشهادة
func (s *CertificateSmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("فشل في قراءة الحالة: %v", err)
	}

	return certJSON != nil, nil
}

// GetAllCertificates عرض قائمة بجميع الشهادات المسجلة
func (s *CertificateSmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var certificates []*Certificate
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
		certificates = append(certificates, &cert)
	}

	return certificates, nil
}
