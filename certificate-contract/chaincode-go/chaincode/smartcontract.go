package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract provides functions for managing Certificates
type SmartContract struct {
	contractapi.Contract
}

// Certificate describes the details of an academic certificate
type Certificate struct {
	CertHash    string `json:"CertHash"`
	Degree      string `json:"Degree"`
	ID          string `json:"ID"`
	IssueDate   string `json:"IssueDate"`
	Issuer      string `json:"Issuer"`
	IsRevoked   bool   `json:"IsRevoked"`
	StudentName string `json:"StudentName"`
}

// getClientMSP returns the MSP ID of the submitting client
func (s *SmartContract) getClientMSP(ctx contractapi.TransactionContextInterface) (string, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to read client MSP: %v", err)
	}
	return mspID, nil
}

// IssueCertificate issues a new certificate to the world state (Org1 only)
func (s *SmartContract) IssueCertificate(
	ctx contractapi.TransactionContextInterface,
	id string, studentName string, degree string, issuer string, issueDate string, certHash string,
) error {
	mspID, err := s.getClientMSP(ctx)
	if err != nil {
		return err
	}
	if mspID != "Org1MSP" {
		return fmt.Errorf("only Org1 can issue certificates")
	}

	exists, err := s.CertificateExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the certificate %s already exists", id)
	}

	cert := Certificate{
		ID:          id,
		StudentName: studentName,
		Degree:      degree,
		Issuer:      issuer,
		IssueDate:   issueDate,
		CertHash:    certHash,
		IsRevoked:   false,
	}
	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// QueryCertificate returns the certificate stored in the world state with given id
func (s *SmartContract) QueryCertificate(ctx contractapi.TransactionContextInterface, id string) (*Certificate, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("the certificate %s does not exist", id)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return nil, err
	}

	return &cert, nil
}

// RevokeCertificate marks a certificate as revoked (Org1 only)
func (s *SmartContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, id string) error {
	mspID, err := s.getClientMSP(ctx)
	if err != nil {
		return err
	}
	if mspID != "Org1MSP" {
		return fmt.Errorf("only Org1 can revoke certificates")
	}

	cert, err := s.QueryCertificate(ctx, id)
	if err != nil {
		return err
	}

	cert.IsRevoked = true

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, certJSON)
}

// VerifyCertificate checks if a certificate exists and is not revoked
func (s *SmartContract) VerifyCertificate(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	cert, err := s.QueryCertificate(ctx, id)
	if err != nil {
		return false, err
	}

	return !cert.IsRevoked, nil
}

// CertificateExists returns true when a certificate with given ID exists in world state
func (s *SmartContract) CertificateExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return certJSON != nil, nil
}

// GetAllCertificates returns all certificates found in world state
func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
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

	return certs, nil
}
