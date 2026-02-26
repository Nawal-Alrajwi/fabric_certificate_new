package chaincode_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/queryresult"
	"github.com/hyperledger/fabric-samples/certificate-contract/chaincode-go/chaincode"
	"github.com/hyperledger/fabric-samples/certificate-contract/chaincode-go/chaincode/mocks"
	"github.com/stretchr/testify/require"
)

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

//go:generate counterfeiter -o mocks/clientidentity.go -fake-name ClientIdentity . clientIdentity
type clientIdentity interface {
	cid.ClientIdentity
}

func TestIssueCertificate(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	clientIdentity := &mocks.ClientIdentity{}
	transactionContext.GetStubReturns(chaincodeStub)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	clientIdentity.GetMSPIDReturns("Org1MSP", nil)

	certContract := chaincode.SmartContract{}
	err := certContract.IssueCertificate(transactionContext, "cert1", "Alice", "BSc", "University", "2024-01-01", "hash123")
	require.NoError(t, err)

	// Test duplicate certificate
	chaincodeStub.GetStateReturns([]byte{}, nil)
	err = certContract.IssueCertificate(transactionContext, "cert1", "Alice", "BSc", "University", "2024-01-01", "hash123")
	require.EqualError(t, err, "the certificate cert1 already exists")

	// Test state error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve certificate"))
	err = certContract.IssueCertificate(transactionContext, "cert1", "Alice", "BSc", "University", "2024-01-01", "hash123")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve certificate")

	// Test unauthorized org
	chaincodeStub.GetStateReturns(nil, nil)
	clientIdentity.GetMSPIDReturns("Org2MSP", nil)
	err = certContract.IssueCertificate(transactionContext, "cert1", "Alice", "BSc", "University", "2024-01-01", "hash123")
	require.EqualError(t, err, "only Org1 can issue certificates")

	// Test MSP error
	clientIdentity.GetMSPIDReturns("", fmt.Errorf("failed to get MSP"))
	err = certContract.IssueCertificate(transactionContext, "cert1", "Alice", "BSc", "University", "2024-01-01", "hash123")
	require.EqualError(t, err, "failed to read client MSP: failed to get MSP")
}

func TestQueryCertificate(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	expectedCert := &chaincode.Certificate{ID: "cert1", StudentName: "Alice"}
	bytes, err := json.Marshal(expectedCert)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	certContract := chaincode.SmartContract{}
	cert, err := certContract.QueryCertificate(transactionContext, "cert1")
	require.NoError(t, err)
	require.Equal(t, expectedCert, cert)

	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve certificate"))
	_, err = certContract.QueryCertificate(transactionContext, "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve certificate")

	chaincodeStub.GetStateReturns(nil, nil)
	cert, err = certContract.QueryCertificate(transactionContext, "cert1")
	require.EqualError(t, err, "the certificate cert1 does not exist")
	require.Nil(t, cert)
}

func TestRevokeCertificate(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	clientIdentity := &mocks.ClientIdentity{}
	transactionContext.GetStubReturns(chaincodeStub)
	transactionContext.GetClientIdentityReturns(clientIdentity)
	clientIdentity.GetMSPIDReturns("Org1MSP", nil)

	expectedCert := &chaincode.Certificate{ID: "cert1"}
	bytes, err := json.Marshal(expectedCert)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	certContract := chaincode.SmartContract{}
	err = certContract.RevokeCertificate(transactionContext, "cert1")
	require.NoError(t, err)

	// Test certificate not found
	chaincodeStub.GetStateReturns(nil, nil)
	err = certContract.RevokeCertificate(transactionContext, "cert1")
	require.EqualError(t, err, "the certificate cert1 does not exist")

	// Test unauthorized org
	clientIdentity.GetMSPIDReturns("Org2MSP", nil)
	err = certContract.RevokeCertificate(transactionContext, "cert1")
	require.EqualError(t, err, "only Org1 can revoke certificates")
}

func TestVerifyCertificate(t *testing.T) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	// Valid certificate
	cert := &chaincode.Certificate{ID: "cert1", IsRevoked: false}
	bytes, err := json.Marshal(cert)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	certContract := chaincode.SmartContract{}
	valid, err := certContract.VerifyCertificate(transactionContext, "cert1")
	require.NoError(t, err)
	require.True(t, valid)

	// Revoked certificate
	cert.IsRevoked = true
	bytes, err = json.Marshal(cert)
	require.NoError(t, err)

	chaincodeStub.GetStateReturns(bytes, nil)
	valid, err = certContract.VerifyCertificate(transactionContext, "cert1")
	require.NoError(t, err)
	require.False(t, valid)

	// Certificate not found
	chaincodeStub.GetStateReturns(nil, nil)
	_, err = certContract.VerifyCertificate(transactionContext, "cert1")
	require.EqualError(t, err, "the certificate cert1 does not exist")
}

func TestGetAllCertificates(t *testing.T) {
	cert := &chaincode.Certificate{ID: "cert1"}
	bytes, err := json.Marshal(cert)
	require.NoError(t, err)

	iterator := &mocks.StateQueryIterator{}
	iterator.HasNextReturnsOnCall(0, true)
	iterator.HasNextReturnsOnCall(1, false)
	iterator.NextReturns(&queryresult.KV{Value: bytes}, nil)

	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)

	chaincodeStub.GetStateByRangeReturns(iterator, nil)
	certContract := &chaincode.SmartContract{}
	certs, err := certContract.GetAllCertificates(transactionContext)
	require.NoError(t, err)
	require.Equal(t, []*chaincode.Certificate{cert}, certs)

	iterator.HasNextReturns(true)
	iterator.NextReturns(nil, fmt.Errorf("failed retrieving next item"))
	certs, err = certContract.GetAllCertificates(transactionContext)
	require.EqualError(t, err, "failed retrieving next item")
	require.Nil(t, certs)

	chaincodeStub.GetStateByRangeReturns(nil, fmt.Errorf("failed retrieving all certificates"))
	certs, err = certContract.GetAllCertificates(transactionContext)
	require.EqualError(t, err, "failed retrieving all certificates")
	require.Nil(t, certs)
}
