package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract defines the structure for our certificate functions
type SmartContract struct {
	contractapi.Contract
}

// Certificate Structure representing the degree data in CouchDB
type Certificate struct {
	CertHash    string `json:"CertHash"`
	Degree      string `json:"Degree"`
	ID          string `json:"ID"`
	IsRevoked   bool   `json:"IsRevoked"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
	StudentName string `json:"StudentName"`
}

// Helper: getClientMSP retrieves the organization ID of the caller
func (s *SmartContract) getClientMSP(ctx contractapi.TransactionContextInterface) (string, error) {
	clientIdentity := ctx.GetClientIdentity()
	mspID, err := clientIdentity.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to read client identity: %v", err)
	}
	return mspID, nil
}

// 1️⃣ IssueCertificate (Org1 Only)
func (s *SmartContract) IssueCertificate(
	ctx contractapi.TransactionContextInterface,
	id string,
	studentName string,
	degree string,
	issuer string,
	certHash string,
	issueDate string) error {

	// RBAC CHECK
	mspID, err := s.getClientMSP(ctx)
	if err != nil || mspID != "Org1MSP" {
		return fmt.Errorf("access denied: only Org1 can issue certificates")
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

// 2️⃣ RevokeCertificate (Org2 Only)
func (s *SmartContract) RevokeCertificate(
	ctx contractapi.TransactionContextInterface,
	id string) error {

	mspID, err := s.getClientMSP(ctx)
	if err != nil || mspID != "Org2MSP" {
		return fmt.Errorf("access denied: only Org2 can revoke certificates")
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

// 3️⃣ VerifyCertificate (Open Read)
func (s *SmartContract) VerifyCertificate(
	ctx contractapi.TransactionContextInterface,
	id string,
	certHash string) (bool, error) {

	cert, err := s.ReadCertificate(ctx, id)
	if err != nil {
		return false, fmt.Errorf("verification failed: certificate not found")
	}

	isValid := cert.CertHash == certHash && !cert.IsRevoked
	return isValid, nil
}

// 4️⃣ ReadCertificate (Helper)
func (s *SmartContract) ReadCertificate(
	ctx contractapi.TransactionContextInterface,
	id string) (*Certificate, error) {

	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
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

// 5️⃣ CertificateExists (Helper)
func (s *SmartContract) CertificateExists(
	ctx contractapi.TransactionContextInterface,
	id string) (bool, error) {

	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return certJSON != nil, nil
}
