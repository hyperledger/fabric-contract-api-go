// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package functionaltests

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/gherkin"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/internal/functionaltests/contracts/complexcontract"
	"github.com/hyperledger/fabric-contract-api-go/internal/functionaltests/contracts/extendedsimplecontract"
	"github.com/hyperledger/fabric-contract-api-go/internal/functionaltests/contracts/simplecontract"
	"github.com/hyperledger/fabric-contract-api-go/internal/functionaltests/contracts/utils"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
	"github.com/hyperledger/fabric-protos-go/peer"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

var contractsMap map[string]contractapi.ContractInterface = map[string]contractapi.ContractInterface{
	"SimpleContract":         new(simplecontract.SimpleContract),
	"ExtendedSimpleContract": NewExtendedContract(),
	"ComplexContract":        NewComplexContract(),
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
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
	lastResponse   peer.Response
	stub           *shimtest.MockStub
	chaincode      *contractapi.ContractChaincode
	metadataFolder string
}

func (sc *suiteContext) cleanup() {
	if sc.metadataFolder != "" {
		os.RemoveAll(sc.metadataFolder)
	}
}

func (sc *suiteContext) createChaincode(name string) error {
	defer sc.cleanup()

	if _, ok := contractsMap[name]; !ok {
		return fmt.Errorf("Invalid contract name %s", name)
	}

	chaincode, err := contractapi.NewChaincode(contractsMap[name])

	if err != nil {
		return fmt.Errorf("expected to get nil for error on create chaincode but got " + err.Error())
	}

	sc.chaincode = chaincode
	sc.stub = shimtest.NewMockStub(name, sc.chaincode)

	return nil
}

func (sc *suiteContext) failCreateChaincode(name string) error {
	err := sc.createChaincode(name)

	if err == nil {
		return fmt.Errorf("Expected to get an error")
	}

	return nil
}

func (sc *suiteContext) createChaincodeMulti(contractsTbl *gherkin.DataTable) error {
	defer sc.cleanup()

	if len(contractsTbl.Rows) > 1 {
		return fmt.Errorf("expected table with one row of contracts")
	}

	contracts := []contractapi.ContractInterface{}

	for _, row := range contractsTbl.Rows {
		for _, cell := range row.Cells {
			contract, ok := contractsMap[cell.Value]

			if !ok {
				return fmt.Errorf("Invalid contract name %s", cell.Value)
			}

			contracts = append(contracts, contract)
		}
	}

	chaincode, err := contractapi.NewChaincode(contracts...)

	if err != nil {
		return fmt.Errorf("expected to get nil for error on create chaincode but got " + err.Error())
	}

	sc.chaincode = chaincode
	sc.stub = shimtest.NewMockStub("MultiContract", sc.chaincode)

	return nil
}

func (sc *suiteContext) createChaincodeAndInit(name string) error {
	err := sc.createChaincode(name)

	if err != nil {
		return err
	}

	return sc.testInitialise()
}

func (sc *suiteContext) setupMetadata(file string) error {
	ex, execErr := os.Executable()
	if execErr != nil {
		return fmt.Errorf("Failed to read metadata from file. Could not find location of executable. %s", execErr.Error())
	}
	exPath := filepath.Dir(ex)
	metadataPath := filepath.Join(exPath, file)

	_, err := os.Stat(metadataPath)

	if os.IsNotExist(err) {
		return errors.New("Failed to read metadata from file. Metadata file does not exist")
	}

	metadataBytes, err := ioutil.ReadFile(metadataPath)

	if err != nil {
		return fmt.Errorf("Failed to read metadata from file. Could not read file %s. %s", metadataPath, err)
	}

	metadataFolder := filepath.Join(exPath, metadata.MetadataFolder)

	os.MkdirAll(metadataFolder, os.ModePerm)
	ioutil.WriteFile(filepath.Join(metadataFolder, metadata.MetadataFile), metadataBytes, os.ModePerm)

	sc.metadataFolder = metadataFolder

	return nil
}

func (sc *suiteContext) testInitialise() error {
	txID := strconv.Itoa(rand.Int())

	sc.stub.MockTransactionStart(txID)
	response := sc.stub.MockInit(txID, [][]byte{})
	sc.stub.MockTransactionEnd(txID)

	if response.GetStatus() != int32(200) {
		return fmt.Errorf("expected to get status 200 on init but got " + strconv.Itoa(int(response.GetStatus())))
	}

	return nil
}

func (sc *suiteContext) invokeChaincode(function string, argsTbl *gherkin.DataTable) error {
	txID := strconv.Itoa(rand.Int())

	argBytes := [][]byte{}
	argBytes = append(argBytes, []byte(function))

	if len(argsTbl.Rows) > 1 {
		return fmt.Errorf("expected zero or one table of args")
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

	return nil
}

func (sc *suiteContext) checkSuccessResponse(result string) error {
	if sc.lastResponse.GetStatus() != int32(200) {
		return fmt.Errorf("expected to get status 200 on invoke %s", sc.lastResponse.GetMessage())
	}

	payload := string(sc.lastResponse.GetPayload())
	if result != "" && payload != result {
		return fmt.Errorf("expected to get payload " + result + " but got " + payload)
	}

	return nil
}

func (sc *suiteContext) checkFailedResponse(result string) error {
	if sc.lastResponse.GetStatus() == int32(200) {
		return fmt.Errorf("expected to not get status 200 on invoke")
	}

	result = fmt.Sprintf("%s", strings.Join(strings.Split(result, "\\n"), "\n"))

	message := sc.lastResponse.GetMessage()
	if result != "" && message != result {
		return fmt.Errorf("expected to get message " + result + " but got " + message)
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	sc := new(suiteContext)

	s.Step(`^I fail to create chaincode from (?:["'](.*?)["'])$`, sc.failCreateChaincode)
	s.Step(`^I have created chaincode from (?:["'](.*?)["'])$`, sc.createChaincode)
	s.Step(`^I have created chaincode from multiple contracts$`, sc.createChaincodeMulti)
	s.Step(`^I have created and initialised chaincode (?:["'](.*?)["'])$`, sc.createChaincodeAndInit)
	s.Step(`^I am using metadata file (?:["'](.*?)["'])$`, sc.setupMetadata)
	s.Step(`^I (?:should\s)?be able to initialise the chaincode`, sc.testInitialise)
	s.Step(`^I submit the (?:"(.*?)") transaction$`, sc.invokeChaincode)
	s.Step(`^I (?:should\s)?receive a successful response\s?(?:(?:["'](.*?)["'])?)$`, sc.checkSuccessResponse)
	s.Step(`^I (?:should\s)?receive an unsuccessful response\s?(?:(?:["'](.*?)["'])?)$`, sc.checkFailedResponse)
}
