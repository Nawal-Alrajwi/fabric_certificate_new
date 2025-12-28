package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract لخدمة إدارة الشهادات الإلكترونية
type SmartContract struct {
	contractapi.Contract
}

// Certificate يمثل هيكل الشهادة الرقمية
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Major       string `json:"Major"`
	University  string `json:"University"`
	IssueDate   string `json:"IssueDate"`
	Grade       string `json:"Grade"`
	IssuerID    string `json:"IssuerID"`
}

// IssueCertificate إصدار شهادة جديدة
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, major string, university string, issueDate string, grade string, issuerID string) error {
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

// QueryCertificate استرجاع بيانات شهادة
func (s *SmartContract) QueryCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
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

// RevokeCertificate تم تصحيح اسم الهيكل هنا
func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("الشهادة %s غير موجودة لإلغائها", id)
	}

	return ctx.GetStub().DelState(id)
}

// CertificateExists التحقق من الوجود
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("فشل في قراءة الحالة: %v", err)
	}

	return certJSON != nil, nil
}

// GetAllCertificates جلب الكل
func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
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
