/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-samples/certificate-contract/chaincode-go/chaincode"
)

func main() {
	certChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating certificate-contract chaincode: %v", err)
	}

	if err := certChaincode.Start(); err != nil {
		log.Panicf("Error starting certificate-contract chaincode: %v", err)
	}
}
