package chaincode

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Certificate تعريف هيكل الشهادة كما قدمته
type Certificate struct {
	CertHash    string `json:"CertHash"`    
	Degree      string `json:"Degree"`      
	ID          string `json:"ID"`          
	IsRevoked   bool   `json:"IsRevoked"`   
	IssueDate   string `json:"IssueDate"`   
	Issuer      string `json:"Issuer"`      
	StudentName string `json:"StudentName"` 
}

// SmartContract defines the structure for our chaincode
type SmartContract struct {
	contractapi.Contract
}

// IssueCertificate إصدار شهادة جديدة وإضافتها إلى World State
func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface, id string, studentName string, degree string, issuer string, issueDate string, certHash string) error {
	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("الشهادة ذات الرقم %s موجودة مسبقاً", id)
	}

	certificate := Certificate{
		ID:          id,
		StudentName: studentName,
		Degree:      degree,
		Issuer:      issuer,
		IssueDate:   issueDate,
		CertHash:    certHash,
		IsRevoked:   false, // الشهادة فعالة عند الإصدار
	}

	certJSON, err := json.Marshal(certificate)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// ReadCertificate قراءة بيانات شهادة معينة باستخدام الرقم التسلسلي
func (s *SmartContract) ReadCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("فشل في القراءة من world state: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("الشهادة %s غير موجودة", id)
	}

	var certificate Certificate
	err = json.Unmarshal(certJSON, &certificate)
	if err != nil {
		return nil, err
	}

	return &certificate, nil
}

// RevokeCertificate إلغاء صلاحية شهادة (بدلاً من الحذف، يفضل تغيير الحالة في الأنظمة الأكاديمية)
func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	certificate, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return err
	}

	certificate.IsRevoked = true
	certJSON, err := json.Marshal(certificate)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// CertificateExists للتأكد من وجود الشهادة
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("فشل في قراءة world state: %v", err)
	}

	return certJSON != nil, nil
}

// GetAllCertificates استرجاع كافة الشهائد المسجلة في النظام
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

		var certificate Certificate
		err = json.Unmarshal(queryResponse.Value, &certificate)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, &certificate)
	}

	return certificates, nil
}
