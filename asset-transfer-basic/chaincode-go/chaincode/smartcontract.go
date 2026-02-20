package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract structure
type SmartContract struct {
	contractapi.Contract
}

// Certificate structure
type Certificate struct {
	ID          string `json:"ID"`
	StudentName string `json:"StudentName"`
	Degree      string `json:"Degree"`
	Issuer      string `json:"Issuer"`
	IssueDate   string `json:"IssueDate"`
	CertHash    string `json:"CertHash"`
	IsRevoked   bool   `json:"IsRevoked"`
}

///////////////////////////////////////////////////////////
// 1️⃣ IssueCertificate
///////////////////////////////////////////////////////////

func (s *SmartContract) IssueCertificate(
	ctx contractapi.TransactionContextInterface,
	id string,
	studentName string,
	degree string,
	issuer string,
	issueDate string,
	certHash string) error {

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

	return ctx.GetStub().PutState(id, certJSON)
}

///////////////////////////////////////////////////////////
// 2️⃣ VerifyCertificate
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

///////////////////////////////////////////////////////////
// Main
///////////////////////////////////////////////////////////

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode: %v", err))
	}

	if err := chaincode.Start(); err != nil {
		panic(fmt.Sprintf("Error starting chaincode: %v", err))
	}
}
