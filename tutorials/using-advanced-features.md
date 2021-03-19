# Using advanced features

## Tutorial contents
- [Prerequisites](#prerequisites)
- [Transaction hooks](#transaction-hooks)
- [Handling unknown function calls](#handling-unknown-function-calls)
- [Chaincode metadata](#chaincode-metadata)
- [What to do next?](#what-to-do-next)

## Prerequisites
This tutorial will assume you have:
- Completed [Getting Started](./getting-started.md)

## Transaction hooks
Creating a chaincode using the contractapi package provides the ability for you to specify functions to be called before and after each call to a contract.

You may have noticed when writing the code in the [previous tutorial](./getting-started.md) that each function performs the same task as its first action, reading from the world state. It would therefore be useful to create one function to do this and set it up to run before each transaction. The transaction context sent to before and after functions is the same instance as the called function receives. We can therefore set data in our before function on this transaction context and use that data in our called function. Likewise we can set data in our called function and use it in an after function. One thing of note is that, since in fabric you [cannot read your own writes](https://hyperledger-fabric.readthedocs.io/en/latest/readwrite.html), if you write to the world state in your before function the called function will not see that updated value.

Before and after functions do not follow the same structure [rules](./getting-started.md#writing-contract-functions) as contract functions. Functions specified to be called before the call cannot take any parameter other than the transaction context and those specified to be called after can only take the transaction context and an interface type. For example:

```
func MyBeforeTransaction(ctx contractapi.TransactionContextInterface) error {
	...
}

func (mc *MyContract) DoSomething(ctx contractapi.TransactionContextInterface) string {
	return "Hello World"
}

func MyAfterTransaction(ctx contractapi.TransactionContextInterface, iface interface{}) error {
	...
}
```

Notice that neither the before or after function in the example above directly receive the parameter data, this is as these functions have to be generic to all calls to the contract. The raw arguments passed in to the call can be accessed using the [stub](https://godoc.org/github.com/hyperledger/fabric-chaincode-go/shim#ChaincodeStub) via the transaction context. The interface value provided to the after function is the value returned by the named function. If we take the above example then that interface for a call to `DoSomething` would be the string `Hello World`.

> Note: if the named function has no defined success response or it returns the type `interface{}` as its success response and has returned nil for that interface the after transaction will receive a nil value for its interface parameter of type `contractapi.UndefinedInterface`. Comparing this value to nil will result in false unless it is typecast.

Both before and after functions can return zero, one or two values although non-error returns are ignored. If the specified before function is defined to return an error and returns a non nil error value when called the named and (if set) after functions are not called and an error is returned to the peer with the before function's returned error value. For example in the following setup if a user were to try and invoke the function `DoSomething` then since the before function returns an error DoSomething is not called and neither is the after function. Instead `Before Failed` would be returned as an error response to the request.

```
func MyBeforeTransaction(ctx contractapi.TransactionContextInterface) error {
	return errors.New("Before failed")
}

func (mc *MyContract) DoSomething(ctx contractapi.TransactionContextInterface) string {
	return "Hello World"
}

func MyAfterTransaction(ctx contractapi.TransactionContextInterface, iface interface{}) error {
	return nil
}
```

Likewise if the named function (e.g. `DoSomething`) errors the after function is not called and the error is returned to the peer. If an after function returns a non nil error then again the peer receives that error. If the after function does not return an error type or the returned error type is nil then the success response from the named function is returned to the peer whether the after function has a success response or not.

> Note: In chaincode when a peer receives an error response anything written to the world state during that call is undone.

To create your function to be called before each transaction create a new file called `transaction-context.go` in the same folder as you used for the [previous tutorial](./getting-started#housekeeping). The following code should be entered in this file. In here we will create our custom transaction context to store data retrieved in our before function for use in the called function. Custom transaction contexts must implement the [contractapi.SettableTransactionContextInterface](https://godoc.org/github.com/hyperledger/fabric-contract-api-go/contractapi#SettableTransactionContextInterface). Like when defining a contract, the easiest way to meet this interface is to embed a struct from the contractapi, this time the standard transaction context.

```
package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// CustomTransactionContext adds methods of storing and retrieving additional data for use
// with before and after transaction hooks
type CustomTransactionContext struct {
	contractapi.TransactionContext
	data []byte
}

// GetData return set data
func (ctc *CustomTransactionContext) GetData() []byte {
	return ctc.data
}

// SetData provide a value for data
func (ctc *CustomTransactionContext) SetData(data []byte) {
	ctc.data = data
}
```

In the previous tutorial it was mentioned that it is more practical from a testing perspective to take an interface for the transaction context. So that we can do this again for our custom transaction context, we must define an interface it meets. So that you can still use the functions from the default transaction context, embed `contractapi.TransactionContextInterface` in the custom transaction interface. Define this interface in the same file as you defined your custom transaction context.

```
// CustomTransactionContextInterface interface to define interaction with custom transaction context
type CustomTransactionContextInterface interface {
	contractapi.TransactionContextInterface
	GetData() []byte
	SetData([]byte)
}
```

Now that you have your custom transaction context interface, you can update each of the functions of `SimpleContract` so that you can use the new functionality. Replace:

```
ctx contractapi.TransactionContextInterface
```

with:

```
ctx CustomTransactionContextInterface
```

The chaincode needs to be informed to use this new transaction context when transactions relate to `SimpleContract`. The chaincode calls `GetTransactionContextHandler()` on the contract to get which transaction context to use. Since `SimpleContract` embeds the `contractapi.Contract` struct you can set the value to be returned when that function is called for your simple contract, by setting an instance of the custom transaction context as the property for `TransactionContextHandler`. Set this in your main function on the `simpleContract` instance before it is used to create the chaincode.

```
simpleContract.TransactionContextHandler = new(CustomTransactionContext)
```

Now that you have a custom transaction context to store the read world state value you can create the function that will be called before each transaction with the simple contract. Create a new file called `utils.go` and add the function to read the world state. As mentioned earlier since before functions do not have parameters (outside of the transaction context) the function needs to use the stub to get the raw transaction arguments. As these are the raw arguments it is important to note that these will not be formatted as they would be for the named function and will still be in their string form. We also cannot rely on the arguments being in a format parsable to the goal type or that there are the correct number of arguments, as both these checks occur after the before function is called.

```
package main

import (
	"errors"
)

// GetWorldState takes the first transaction arg as the key and sets
// what is found in the world state for that key in the transaction context
func GetWorldState(ctx CustomTransactionContextInterface) error {
	_, params := ctx.GetStub().GetFunctionAndParameters()

	if len(params) < 1 {
		return errors.New("Missing key for world state")
	}

	existing, err := ctx.GetStub().GetState(params[0])

	if err != nil {
		return errors.New("Unable to interact with world state")
	}

	ctx.SetData(existing)

	return nil
}
```

Since all the functions in simple contract take the key as their first parameter, the function uses the first parameter from the raw transaction arguments as the key for interacting with the world state and writes the world state response to the transaction context so that it can be accessed in the main function call. Should the reading of the world state fail (note: it won't fail for a missing key) the function will return an error meaning that the rest of the transaction call does not happen and the error is returned to the peer.

Like with the custom transaction context, the chaincode needs to be informed that we would like to use an advanced feature. The chaincode calls `GetBeforeTransaction()` on the contract to determine whether there is a function to call before the named function passed as part of the transaction. Just like with setting of the custom transaction context, because `SimpleContract` embeds the `contractapi.Contract` struct we can just set a property of the `SimpleContract` instance used in the chaincode creation in `main`.

```
simpleContract.BeforeTransaction = GetWorldState
```

As the `GetWorldState` function is now set to be called before each transaction relating to `SimpleContract`, you can remove the repetitious code at the start of each of the functions of `SimpleContract`. Replace:

```
existing, err := ctx.GetStub().GetState(key)

if err != nil {
    return errors.New("Unable to interact with world state")
}
```

with:
```
existing := ctx.GetData()
```

> Note: As you have removed the definition of err you will need to change err = to err := in the `Create` and `Update` functions

Your chaincode should now work exactly the same as before. If you have torn down your network from the previous tutorial, you can run through the steps outlined [here](./getting-started.md#testing-your-chaincode-as-a-developer) to bring your chaincode up and interact with it. Otherwise stop the chaincode process, rebuild your go program and restart it using the same command. You can then interact with the same invoke/query commands.

> Note: if you have not torn down the network data will still persist in the world state and therefore you will need to use another key (e.g. `KEY_2`) in your commands

## Handling unknown function calls

By default if a function name is passed during an instantiate, invoke or query request that is unknown to the chaincode the chaincode returns an error response to the peer to let the user know of the issue. For example, when a user misspells a known function or enters a non-existent one. To see this in action issue the following command:

```
peer chaincode query -n mycc -c '{"Args":["BadFunction", "KEY_1"]}' -C myc
```

It is possible, however, to specify a custom handler for these unknown function requests for each contract. A handler for unknown requests may (optionally) take the transaction context as its sole parameter. It does not need to be public or a function of the contract. The unknown transaction handler may return an error type, if it does return a value for this error then any after transaction specified for the contract will not be run. Any before transaction function is always run.

Define your own function for handling unknown function names in transactions. In the `utils.go` file (where you defined `GetWorldState`) create the following function:

```
// UnknownTransactionHandler returns a shim error
// with details of a bad transaction request
func UnknownTransactionHandler(ctx CustomTransactionContextInterface) error {
	fcn, args := ctx.GetStub().GetFunctionAndParameters()
	return fmt.Errorf("Invalid function %s passed with args %v", fcn, args)
}
```

> Note: if your editor doesn't do it for you automatically, add fmt to the list of imports in `utils.go`

This function will return an error whenever it is called providing more detail than the default unknown response would. Although it doesn't interact with the world state, it still takes the context to allow it access to the details of the transaction.

The process for setting the unknown transaction handler to be used by the chaincode follows a similar path to setting a function to be called before each transaction. The chaincode calls the function `GetUnknownTransaction()` on contracts to determine which function to use in the case of an unknown function name being sent for that contract as part of the transaction (or whether to use the default). Since `SimpleContract` embeds the `contractapi.Contract` struct, you can set your custom unknown transaction handler by setting the property `UnknownTransaction` of the `SimpleContract` instance used to create the chaincode in the `main` function.

```
simpleContract.UnknownTransaction = UnknownTransactionHandler
```

If in the chaincode docker container terminal you now stop the chaincode process, rebuild the go program and then restart the chaincode process then in the CLI docker container terminal run the following query you should see that `UnknownTransactionHandler` is called for bad function names:

```
peer chaincode query -n mycc -c '{"Args":["BadFunction", "KEY_1"]}' -C myc
```

Notice that the output differs from what was returned when you issued the same command before setting up the custom unknown transaction handler.

## Chaincode metadata
Chaincode created using the contractapi package automatically has generated for it a system contract which provides metadata about the chaincode. This metadata describes the contracts that form the chaincode, describing their functions, the parameters those functions take, as well as function return values. The metadata produced follows this [schema](https://raw.githubusercontent.com/hyperledger/fabric-contract-api-go/main/metadata/schema/schema.json).

In Go the metadata is produced automatically for you using reflection, due to limitations of Go reflection the parameter names of functions in the metadata will not match the chaincode code but will instead use param0, param1, ..., paramN.

You can view the metadata for your chaincode by querying the system chaincode 'org.hyperledger.fabric' and its function 'GetMetadata'. This is also how you view the metadata of chaincodes written in Node and Java (as long as they follow the [Fabric programming model](https://hyperledger-fabric.readthedocs.io/en/release-1.4/whatsnew.html#improved-programming-model-for-developing-applications)).

To see the metadata of the chaincode made in this tutorial issue the following command in the CLI docker terminal:

```
peer chaincode query -n mycc -c '{"Args":["org.hyperledger.fabric:GetMetadata"]}' -C myc
```

Notice in the chaincode that contract functions have a property `tag` which contains a string array with one element `submit`. A submit tag is used in the metadata to indicate that the function when called as part of a transaction should be "submitted" rather than "evaluated" by the client. Submitting a transaction (invoking) means that data can be written to the world state, evaluating (querying) runs the function as read only. The `Read` function of `SimpleContract` is tagged as `submit` however in the code it never writes to the world state. This is as all contract functions are tagged as `submit` by default when the contractapi package is used to create chaincode. You must therefore explicitly mark which functions are for evaluating rather than submitting. It is important to note that metadata merely provides a guide for interaction with the chaincode and its contracts. Having a tag of submit or evaluate does not force the client issuing the transaction to use that method. A function marked as evaluate can be called via a submit transaction.

Chaincode created using the contractapi calls the `GetEvaluateTransactions()` function of a contract (should it exist) to retrieve a list of name of functions that are evaluate rather than submit. Define a `GetEvaluateTransactions` function on `SimpleContract` and have it return `Read` as the only element of the string array:

```
// GetEvaluateTransactions returns functions of SimpleContract not to be tagged as submit
func (sc *SimpleContract) GetEvaluateTransactions() []string {
	return []string{"Read"}
}
```

If, in the CLI docker terminal, you kill the chaincode process, rebuild the chaincode, restart it and then issue the metadata query command again, you should now see that the `Read` function is no longer tagged as `submit`.

The info section of metadata is by default filled with a title and version. This section exists for each contract and the chaincode as a whole. If no version is set then this defaults to "latest". If no title is set for a chaincode then "undefined" is used in its info section, if no title is set for a contract then the struct name is used in its info section. You can set the info section of metadata for chaincode and contract respectively by setting the `Info` property of the chaincode and contract instances used in `main`.

> Note: the version in the chaincode's info section is not linked directly to the version used when creating the chaincode in the network

## What to do next?
Follow the [Managing objects](./managing-objects.md) tutorial.
