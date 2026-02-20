package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/pkg/statebased"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Certificate structure
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
// 🔐 Helper: Get MSP ID (RBAC)
///////////////////////////////////////////////////////////

func (s *SmartContract) getClientMSP(ctx contractapi.TransactionContextInterface) (string, error) {
	return ctx.GetClientIdentity().GetMSPID()
}

///////////////////////////////////////////////////////////
// 1️⃣ IssueCertificate (RBAC + SBE)
///////////////////////////////////////////////////////////

func (s *SmartContract) IssueCertificate(
	ctx contractapi.TransactionContextInterface,
	id string,
	studentName string,
	degree string,
	issuer string,
	issueDate string,
	certHash string) error {

	// 🔐 RBAC (Org-Level Only)
	mspID, err := s.getClientMSP(ctx)
	if err != nil {
		return err
	}
	if mspID != "Org1MSP" {
		return fmt.Errorf("only Org1 can issue certificates")
	}

	// Check if certificate already exists
	existing, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read world state: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("certificate %s already exists", id)
	}

	certificate := Certificate{
		ID:          id,
		StudentName: studentName,
		Degree:      degree,
		Issuer:      issuer,
		IssueDate:   issueDate,
		CertHash:    certHash,
		IsRevoked:   false,
	}

	certJSON, err := json.Marshal(certificate)
	if err != nil {
		return err
	}

	// Store in public world state
	err = ctx.GetStub().PutState(id, certJSON)
	if err != nil {
		return err
	}

	// 🔐 State-Based Endorsement (Only Org1 can modify)
	ep, err := statebased.NewStateEP(nil)
	if err != nil {
		return err
	}

	err = ep.AddOrgs(statebased.RoleTypePeer, "Org1MSP")
	if err != nil {
		return err
	}

	policy, err := ep.Policy()
	if err != nil {
		return err
	}

	return ctx.GetStub().SetStateValidationParameter(id, policy)
}

///////////////////////////////////////////////////////////
// 2️⃣ VerifyCertificate (Read-Only)
///////////////////////////////////////////////////////////

func (s *SmartContract) VerifyCertificate(
	ctx contractapi.TransactionContextInterface,
	id string,
	providedHash string) (bool, error) {

	certJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read world state: %v", err)
	}
	if certJSON == nil {
		return false, fmt.Errorf("certificate %s does not exist", id)
	}

	var certificate Certificate
	err = json.Unmarshal(certJSON, &certificate)
	if err != nil {
		return false, err
	}

	if certificate.IsRevoked {
		return false, nil
	}

	return certificate.CertHash == providedHash, nil
}
