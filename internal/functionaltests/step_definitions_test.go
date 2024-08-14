// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fvtests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/v2/internal/functionaltests/contracts/complexcontract"
	"github.com/hyperledger/fabric-contract-api-go/v2/internal/functionaltests/contracts/extendedsimplecontract"
	"github.com/hyperledger/fabric-contract-api-go/v2/internal/functionaltests/contracts/simplecontract"
	"github.com/hyperledger/fabric-contract-api-go/v2/internal/functionaltests/contracts/utils"
	"github.com/hyperledger/fabric-contract-api-go/v2/metadata"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

var contractsMap map[string]contractapi.ContractInterface = map[string]contractapi.ContractInterface{
	"SimpleContract":         new(simplecontract.SimpleContract),
	"ExtendedSimpleContract": NewExtendedContract(),
	"ComplexContract":        NewComplexContract(),
}

func NewExtendedContract() *extendedsimplecontract.ExtendedSimpleContract {
	esc := new(extendedsimplecontract.ExtendedSimpleContract)

	esc.TransactionContextHandler = new(utils.CustomTransactionContext)
	esc.BeforeTransaction = utils.GetWorldState
	esc.UnknownTransaction = utils.UnknownTransactionHandler

	return esc
}

func NewComplexContract() *complexcontract.ComplexContract {
	cc := new(complexcontract.ComplexContract)

	cc.TransactionContextHandler = new(utils.CustomTransactionContext)
	cc.BeforeTransaction = utils.GetWorldState
	cc.UnknownTransaction = utils.UnknownTransactionHandler

	return cc
}

type suiteContext struct {
	lastResponse   *peer.Response
	stub           *MockStub
	chaincode      *contractapi.ContractChaincode
	metadataFolder string
}

type suiteContextKey struct{}

func cleanup(ctx context.Context) {
	if sc, ok := ctx.Value(suiteContextKey{}).(suiteContext); ok && sc.metadataFolder != "" {
		os.RemoveAll(sc.metadataFolder)
	}
}

func iHaveCreatedChaincodeFrom(ctx context.Context, name string) (context.Context, error) {
	defer cleanup(ctx)

	if _, ok := contractsMap[name]; !ok {
		return ctx, fmt.Errorf("Invalid contract name %s", name)
	}

	chaincode, err := contractapi.NewChaincode(contractsMap[name])
	if err != nil {
		return ctx, fmt.Errorf("expected to get nil for error on create chaincode but got %s", err.Error())
	}

	sc := suiteContext{}
	sc.chaincode = chaincode
	sc.stub = NewMockStub(sc.chaincode)

	return context.WithValue(ctx, suiteContextKey{}, sc), nil
}

func iHaveCreatedChaincodeFromMultipleContracts(ctx context.Context, contractsTbl *godog.Table) (context.Context, error) {
	defer cleanup(ctx)
	if len(contractsTbl.Rows) > 1 {
		return ctx, fmt.Errorf("expected table with one row of contracts")
	}

	contracts := []contractapi.ContractInterface{}

	for _, row := range contractsTbl.Rows {
		for _, cell := range row.Cells {
			contract, ok := contractsMap[cell.Value]

			if !ok {
				return ctx, fmt.Errorf("Invalid contract name %s", cell.Value)
			}

			contracts = append(contracts, contract)
		}
	}

	chaincode, err := contractapi.NewChaincode(contracts...)

	if err != nil {
		return ctx, fmt.Errorf("expected to get nil for error on create chaincode but got %s", err.Error())
	}
	sc := suiteContext{}
	sc.chaincode = chaincode
	sc.stub = NewMockStub(sc.chaincode)
	return context.WithValue(ctx, suiteContextKey{}, sc), nil
}

func iShouldBeAbleToInitialiseTheChaincode(ctx context.Context) (context.Context, error) {
	sc, ok := ctx.Value(suiteContextKey{}).(suiteContext)
	if !ok {
		return ctx, errors.New("there are no contracts available")
	}

	txID := strconv.Itoa(rand.Int()) //nolint:gosec

	sc.stub.MockTransactionStart(txID)
	response := sc.stub.MockInit(txID, [][]byte{})
	sc.stub.MockTransactionEnd(txID)

	if response.GetStatus() != int32(200) {
		return ctx, fmt.Errorf("expected to get status 200 on init but got %s", strconv.Itoa(int(response.GetStatus())))
	}

	return context.WithValue(ctx, suiteContextKey{}, sc), nil
}

func iShouldReceiveASuccessfulResponse(ctx context.Context, result string) (context.Context, error) {
	sc, ok := ctx.Value(suiteContextKey{}).(suiteContext)
	if !ok {
		return ctx, errors.New("there are no contracts available")
	}

	payload := string(sc.lastResponse.GetPayload())
	if result != "" && payload != result {
		return ctx, fmt.Errorf(`expected to get payload "%s" but got "%s"`, result, payload)
	}

	return ctx, nil

}

func iSubmitTheTransaction(ctx context.Context, function string, argsTbl *godog.Table) (context.Context, error) {
	sc, ok := ctx.Value(suiteContextKey{}).(suiteContext)
	if !ok {
		return ctx, errors.New("there are no contracts available")
	}

	txID := strconv.Itoa(rand.Int()) //nolint:gosec

	argBytes := [][]byte{}
	argBytes = append(argBytes, []byte(function))

	if len(argsTbl.Rows) > 1 {
		return ctx, fmt.Errorf("expected zero or one table of args")
	}

	for _, row := range argsTbl.Rows {
		for _, cell := range row.Cells {
			argBytes = append(argBytes, []byte(cell.Value))
		}
	}

	sc.stub.MockTransactionStart(txID)
	response := sc.stub.MockInvoke(txID, argBytes)
	sc.stub.MockTransactionEnd(txID)

	sc.lastResponse = response

	return context.WithValue(ctx, suiteContextKey{}, sc), nil

}

func iAmUsingMetadataFile(ctx context.Context, file string) (context.Context, error) {
	ex, execErr := os.Executable()
	if execErr != nil {
		return ctx, fmt.Errorf("failed to read metadata from file. Could not find location of executable. %s", execErr.Error())
	}
	exPath := filepath.Dir(ex)
	metadataPath := filepath.Join(exPath, file)

	metadataBytes, err := os.ReadFile(metadataPath)
	if err != nil {
		return ctx, fmt.Errorf("failed to read metadata from file. Could not read file %s. %s", metadataPath, err)
	}

	metadataFolder := filepath.Join(exPath, metadata.MetadataFolder)

	if err := os.MkdirAll(metadataFolder, os.ModePerm); err != nil {
		return ctx, err
	}

	if err := os.WriteFile(filepath.Join(metadataFolder, metadata.MetadataFile), metadataBytes, os.ModePerm); err != nil { //nolint:gosec
		return ctx, err
	}

	sc := suiteContext{}
	sc.metadataFolder = metadataFolder

	return context.WithValue(ctx, suiteContextKey{}, sc), nil

}

func iFailToCreateChaincodeFrom(ctx context.Context, name string) (context.Context, error) {
	_, err := iHaveCreatedChaincodeFrom(ctx, name)

	if err == nil {
		return ctx, fmt.Errorf("Expected to get an error")
	}

	return ctx, nil
}

func iShouldReceiveAnUnsuccessfulResponse(ctx context.Context, result string) (context.Context, error) {

	sc, ok := ctx.Value(suiteContextKey{}).(suiteContext)
	if !ok {
		return ctx, errors.New("there are no contracts available")
	}

	if sc.lastResponse.GetStatus() == int32(200) {
		return ctx, fmt.Errorf("expected to not get status 200 on invoke")
	}

	result = strings.Join(strings.Split(result, "\\n"), "\n")

	message := sc.lastResponse.GetMessage()
	if result != "" && message != result {
		return ctx, fmt.Errorf(`expected to get message "%s" but got "%s"`, result, message)
	}
	return ctx, nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^I have created chaincode from "([^"]*)"$`, iHaveCreatedChaincodeFrom)
	ctx.Step(`^I have created chaincode from multiple contracts$`, iHaveCreatedChaincodeFromMultipleContracts)
	ctx.Step(`^I should be able to initialise the chaincode$`, iShouldBeAbleToInitialiseTheChaincode)
	ctx.Step(`^I have initialised the chaincode$`, iShouldBeAbleToInitialiseTheChaincode)
	ctx.Step(`^I (?:should\s)?receive a successful response\s?(?:(?:["'](.*?)["'])?)$`, iShouldReceiveASuccessfulResponse)
	ctx.Step(`^I submit the "([^"]*)" transaction$`, iSubmitTheTransaction)
	ctx.Step(`^I am using metadata file "([^"]*)"$`, iAmUsingMetadataFile)
	ctx.Step(`^I fail to create chaincode from "([^"]*)"$`, iFailToCreateChaincodeFrom)
	ctx.Step(`^I should receive an unsuccessful response "([^"]*)"$`, iShouldReceiveAnUnsuccessfulResponse)
}
