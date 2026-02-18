package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

///////////////////////////////////////////////////////////
// Certificate Structure
///////////////////////////////////////////////////////////

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
		return "", fmt.Errorf("failed to read client identity: %v", err)
	}

	return mspID, nil
}

///////////////////////////////////////////////////////////
// 1️⃣ IssueCertificate (Org1 Only)
///////////////////////////////////////////////////////////

func (s *SmartContract) IssueCertificate(
	ctx contractapi.TransactionContextInterface,
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
		return fmt.Errorf("access denied: only Org1 can issue certificates")
	}
	// -------------------

	if id == "" || studentName == "" || degree == "" || issuer == "" || certHash == "" || issueDate == "" {
		return fmt.Errorf("all fields are required")
	}

	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("certificate %s already exists", id)
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

func (s *SmartContract) RevokeCertificate(
	ctx contractapi.TransactionContextInterface,
	id string) error {

	// --- RBAC CHECK ---
	mspID, err := s.getClientMSP(ctx)
	if err != nil {
		return err
	}

	if mspID != "Org2MSP" {
		return fmt.Errorf("access denied: only Org2 can revoke certificates")
	}
	// -------------------

	if id == "" {
		return fmt.Errorf("certificate ID is required")
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

func (s *SmartContract) VerifyCertificate(
	ctx contractapi.TransactionContextInterface,
	id string,
	certHash string) (bool, error) {

	if id == "" || certHash == "" {
		return false, fmt.Errorf("certificate ID and hash are required")
	}

	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return false, nil
	}

	isValid := cert.CertHash == certHash && !cert.IsRevoked

	return isValid, nil
}

///////////////////////////////////////////////////////////
// Helper Functions
///////////////////////////////////////////////////////////

func (s *SmartContract) ReadCertificate(
	ctx contractapi.TransactionContextInterface,
	id string) (*Certificate, error) {

	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}

	if certJSON == nil {
		return nil, fmt.Errorf("certificate %s does not exist", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

func (s *SmartContract) CertificateExists(
	ctx contractapi.TransactionContextInterface,
	id string) (bool, error) {

	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}

	return certJSON != nil, nil
}
