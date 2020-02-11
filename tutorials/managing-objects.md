# Managing objects

## Tutorial contents
- [Prerequisites](#prerequisites)
- [Defining an object](#defining-an-object)
- [Building a contract to handle an object](#building-a-contract-to-handle-an-object)
- [Adding a second contract to the chaincode](#adding-a-second-contract-to-the-chaincode)
- [Using a custom name for your contracts](#using-a-custom-name-for-your-contracts)
- [What to do next?](#what-to-do-next)

## Prerequisites
This tutorial will assume you have:
- Completed [Using advanced features](./using-advanced-features.md)

## Defining an object
The chaincode written so far contains a single contract and purely works with taking and returning string values. As mentioned in the [first tutorial](./getting-started.md) functions can take and return many types including structs (and pointers to structs). This tutorial will create a contract which handles the management of an object to show how the contract API handles the taking and returning of non-string types.

Create a new file in your `contract-tutorial` folder called `basic-asset.go`. In here we will create the object to manage, in this case we will call it `BasicAsset`. 

```
package main

// Owner contains the full name of an owner of a basic asset
type Owner struct {
	Forename string `json:"forename"`
	Surname  string `json:"surname"`
}

// BasicAsset a basic asset
type BasicAsset struct {
	ID        string `json:"id"`
	Owner     Owner  `json:"owner"`
	Value     int    `json:"value"`
	Condition int    `json:"condition"`
}

// SetConditionNew set the condition of the asset to mark as new
func (ba *BasicAsset) SetConditionNew() {
	ba.Condition = 0
}

// SetConditionUsed set the condition of the asset to mark as used
func (ba *BasicAsset) SetConditionUsed() {
	ba.Condition = 1
}
```

Notice that the struct properties are tagged with JSON tags. When a call is made to chaincode created using the contractapi package the transaction arguments and returned values are strings which get converted to and from their go value by a serializer. By default this serializer is a JSON serializer which is built on top of the standard JSON marshalling/unmarshalling in Go, it is also possible to use a serializer of your own definition. These tags are therefore used to tell the serializer how to convert to and from the object. In this case it says the property 'ID' is referenced in a JSON string by the property 'id'. These JSON tags are also used in the metadata to describe the object, more detail can be found in the [godoc](https://godoc.org/github.com/hyperledger/fabric-contract-api-go/metadata#GetSchema). This is as the metadata is intended to tell a user of the smart contract what they need to send in a transaction and what to expect in return.

> Note: as the default serializer is built on top of the standard JSON marshalling/unmarshalling in Go, it is possible to write your own handler for the marshalling by creating MarshalJSON and UnmarshalJSON functions. Beware you may need to make use of the `metadata` tag in your struct to ensure that the contract metadata matches this custom setup.

## Building a contract to handle an object

Now that the object is defined, create a new contract to handle it. This contract will handle the business logic of managing our basic asset. This contract can be created in the same way as the simple contract was. Start by creating a new file `complex-contract.go` and add a struct `ComplexContract` which embeds the `contractapi.Contract` struct.

```
package main

import (
    "encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ComplexContract contract for handling BasicAssets
type ComplexContract struct {
	contractapi.Contract
}
```

Now add the first function for the management of the asset, this will be a function to create a new instance of the asset and record this in the world state using the ID as the key. The function, as well as all others in this contract, will perform a get of the passed ID from the world state. Therefore it makes sense to use the same process as in the simple contract, a before function which calls get using the passed ID. This before function will be the same as used by the simple contract, although we could have written a custom function just for this contract. Like with the simple contract, it is necessary to alert the contract API to the use of this function, but that will be set later. As the get function uses a custom context `utils.CustomTransactionContext` the new asset function (and all others in this contract) will use the same transaction context type.

```
// NewAsset adds a new basic asset to the world state using id as key
func (s *ComplexContract) NewAsset(ctx CustomTransactionContextInterface, id string, owner Owner, value int) error {
	existing := ctx.GetData()

	if existing != nil {
		return fmt.Errorf("Cannot create new basic asset in world state as key %s already exists", id)
	}

	ba := new(BasicAsset)
	ba.ID = id
	ba.Owner = owner
	ba.Value = value
	ba.SetConditionNew()

	baBytes, _ := json.Marshal(ba)

	err := ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}
```

The next functions will handle updating the asset in the world state. The first function will update the owner by simply replacing the owner value and the second will update the value by adding the value passed. The change of ownership will also mark the asset as used. Both functions will take the data from the world state and convert it back to a `BasicAsset` before updating the values.

```
// UpdateOwner changes the ownership of a basic asset and mark it as used
func (cc *ComplexContract) UpdateOwner(ctx CustomTransactionContextInterface, id string, newOwner Owner) error {
	existing := ctx.GetData()

	if existing == nil {
		return fmt.Errorf("Cannot update asset in world state as key %s does not exist", id)
	}

	ba := new(BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	ba.Owner = newOwner
	ba.SetConditionUsed()

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}

// UpdateValue changes the value of a basic asset to add the value passed
func (cc *ComplexContract) UpdateValue(ctx CustomTransactionContextInterface, id string, valueAdd int) error {
	existing := ctx.GetData()

	if existing == nil {
		return fmt.Errorf("Cannot update asset in world state as key %s does not exist", id)
	}

	ba := new(BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	ba.Value += valueAdd

	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	return nil
}
```

Add a further function to allow a transaction caller to read your asset from the world state.

```
// GetAsset returns the basic asset with id given from the world state
func (cc *ComplexContract) GetAsset(ctx CustomTransactionContextInterface, id string) (*BasicAsset, error) {
	existing := ctx.GetData()

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(BasicAsset)

	err := json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type BasicAsset", id)
	}

	return ba, nil
}
```

Finally define a `GetEvaluateTransactions` functions for your new contract. Make this function return `GetAsset` since this is the only function that does not write to the world state.

```
// GetEvaluateTransactions returns functions of ComplexContract not to be tagged as submit
func (cc *ComplexContract) GetEvaluateTransactions() []string {
	return []string{"GetAsset"}
}
```

## Adding a second contract to the chaincode

Your `main.go` file will already contain the code to use the simple contract inside chaincode, here you now need to add code to use your complex contract inside the chaincode as well. This is done using the exact same method as the simple contract by creating a new instance of the `ComplexContract` struct and passing this new instance as an argument to the `contractapi.NewChaincode` function. Earlier in this tutorial you programmed the complex contract to make use of the custom transaction context and rely on `GetWorldState` being called before each transaction. Like with the simple contract you must let the chaincode know to use these. Your main function should therefore look like this:

```
func main() {
	simpleContract := new(SimpleContract)
	simpleContract.TransactionContextHandler = new(CustomTransactionContext)
	simpleContract.BeforeTransaction = GetWorldState
	simpleContract.UnknownTransaction = UnknownTransactionHandler

	complexContract := new(ComplexContract)
	complexContract.TransactionContextHandler = new(CustomTransactionContext)
	complexContract.BeforeTransaction = GetWorldState

	cc, err := contractapi.NewChaincode(simpleContract, complexContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
```

You now have a chaincode consisting of two contracts.

> Note: since both contracts are part of the same chaincode they can read and write to the same keys in the ledger.

## Using a custom name for your contracts
When there was only one contract in the chaincode, calling it consisted of just passing the function name. With multiple it is now necessary to know the contract name and namespace calls in the format `<CONTRACT_NAME>:<FUNCTION_NAME>`. By default the contracts can be referenced by their struct name, in this case they would be `SimpleContract` and `ComplexContract`. The first contract passed to the `NewChaincode` function is also the default contract and therefore its functions can be called without namespacing. Sometimes it is desirable to to use custom names for contracts. The chaincode calls `GetName` on the contract when it is created to determine the name of the contract and since both our contracts embed the `contractapi.Contract` struct, we can set the value to be returned from this function by setting the name property of the contracts before calling `NewChaincode` in your `main` function.

```
simpleContract.Name = "org.example.com.SimpleContract"

complexContract.Name = "org.example.com.ComplexContract"
```

You can then set the default contract to be the complex contract by, in your `main` function, setting the `DefaultContract` property of the chaincode to be the name of the complex chaincode. Do this after checking the error value returned by `NewChaincode` or you may get a runtime error as if an error occurs chaincode will be `nil`.

```
cc.DefaultContract = complexContract.GetName()
```

If you have torn down your network from the previous tutorials you can run through the steps outlined [here](./getting-started.md#testing-your-chaincode-as-a-developer) to bring your chaincode up, install and instantiate it (note: since we changed the default contract the invoke and query commands will no longer work). Otherwise, in the chaincode docker container terminal kill the chaincode process, rebuild your go program and then restart it. Issue the following commands to interact with the chaincode:

### Simple contract

```
peer chaincode invoke -n mycc -c '{"Args":["org.example.com.SimpleContract:Create", "KEY_3", "VALUE_1"]}' -C myc

peer chaincode invoke -n mycc -c '{"Args":["org.example.com.SimpleContract:Update", "KEY_3", "VALUE_2"]}' -C myc

peer chaincode query -n mycc -c '{"Args":["org.example.com.SimpleContract:Read", "KEY_3"]}' -C myc
```

### Complex contract

> You can call the complex contract both using its name or by just passing the name of its functions since it is now the default

```
# call without passing name
peer chaincode invoke -n mycc -c '{"Args":["NewAsset", "ASSET_1", "{\"forename\": \"blue\", \"surname\": \"conga\"}", "100"]}' -C myc

# call passing name
peer chaincode invoke -n mycc -c '{"Args":["org.example.com.ComplexContract:UpdateOwner", "ASSET_1", "{\"forename\": \"green\", \"surname\": \"conga\"}"]}' -C myc

peer chaincode invoke -n mycc -c '{"Args":["org.example.com.ComplexContract:UpdateValue", "ASSET_1", "300"]}' -C myc

peer chaincode query -n mycc -c '{"Args":["org.example.com.ComplexContract:GetAsset", "ASSET_1"]}' -C myc
```

## What to do next?
This is the last of the three tutorials in this repo. Next you could extend the contracts developed here, create your own chaincode or play with [fabcar](https://github.com/hyperledger/fabric-samples).