package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Certificate تعريف هيكل الشهادة
type Certificate struct {
	CertHash    string `json:"CertHash"`
	Degree      string `json:"Degree"`
	ID          string `json:"ID"`
	IsRevoked   bool   `json:"IsRevoked"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
	StudentName string `json:"StudentName"`
}

///////////////////////////////////////////////////////////
// 🔐 MSP-Based RBAC Helper
///////////////////////////////////////////////////////////

func (s *SmartContract) getClientMSP(ctx contractapi.TransactionContextInterface) (string, error) {
	clientIdentity := ctx.GetClientIdentity()

	mspID, err := clientIdentity.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("فشل في قراءة هوية العميل: %v", err)
	}

	return mspID, nil
}

///////////////////////////////////////////////////////////
// 1️⃣ IssueCertificate (Org1 Only)
///////////////////////////////////////////////////////////

func (s *SmartContract) IssueCertificate(ctx contractapi.TransactionContextInterface,
	id string,
	studentName string,
	degree string,
	issuer string,
	certHash string,
	issueDate string) error {

	// --- RBAC CHECK ---
	mspID, err := s.getClientMSP(ctx)
	if err != nil {
		return err
	}

	if mspID != "Org1MSP" {
		return fmt.Errorf("غير مصرح لك بإصدار شهادة")
	}
	// -------------------

	// Validation
	if id == "" || studentName == "" || degree == "" || issuer == "" || certHash == "" || issueDate == "" {
		return fmt.Errorf("جميع الحقول مطلوبة")
	}

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
		Degree:      degree,
		Issuer:      issuer,
		CertHash:    certHash,
		IssueDate:   issueDate,
		IsRevoked:   false,
	}

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

///////////////////////////////////////////////////////////
// 2️⃣ QueryAllCertificates (Open Read)
///////////////////////////////////////////////////////////

func (s *SmartContract) QueryAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {

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

///////////////////////////////////////////////////////////
// 3️⃣ RevokeCertificate (Org2 Only)
///////////////////////////////////////////////////////////

func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {

	// --- RBAC CHECK ---
	mspID, err := s.getClientMSP(ctx)
	if err != nil {
		return err
	}

	if mspID != "Org2MSP" {
		return fmt.Errorf("غير مصرح لك بإلغاء الشهادة")
	}
	// -------------------

	if id == "" {
		return fmt.Errorf("معرف الشهادة مطلوب")
	}

	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return err
	}

	if cert.IsRevoked {
		return nil
	}

	cert.IsRevoked = true

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

///////////////////////////////////////////////////////////
// 4️⃣ VerifyCertificate (Open Read)
///////////////////////////////////////////////////////////

func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface,
	id string,
	certHash string) (bool, error) {

	if id == "" || certHash == "" {
		return false, fmt.Errorf("المعرف والبصمة مطلوبة")
	}

	cert, err := s.ReadCe
