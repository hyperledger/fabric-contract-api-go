# Getting started

## Tutorial contents
- [Prerequisites](#prerequisites)
- [Housekeeping](#housekeeping)
- [Declaring a contract](#declaring-a-contract)
- [Writing contract functions](#writing-contract-functions)
- [Using contracts in chaincode](#using-contracts-in-chaincode)
- [Testing your chaincode as a developer](#testing-your-chaincode-as-a-developer)
- [What to do next?](#what-to-do-next)

## Prerequisites
This tutorial will assume you have:
- A clone of [fabric-samples](https://github.com/hyperledger/fabric-samples)
- [Go 1.13.x](https://golang.org/doc/install)
- [Docker](https://docs.docker.com/install/)
- [Docker compose](https://docs.docker.com/compose/install/)

## Housekeeping
Since this tutorial will make use of fabric-samples' `chaincode-docker-devmode` setup you should be developing within `fabric-samples/chaincode`. Make a folder inside `fabric-samples/chaincode` called `contract-tutorial` and open your preferred editor there. In your terminal run the command

```
go mod init github.com/hyperledger/fabric-samples/chaincode/contract-tutorial
```

to setup go modules. You can then run
 
```
go get -u github.com/hyperledger/fabric-contract-api-go
```

to get the latest release of fabric-contract-api-go for use in your chaincode.

## Declaring a contract
The contractapi generates chaincode by taking one or more "contracts" that it bundles into a running chaincode. The first thing we will do here is declare a contract for use in our chaincode. This contract will be simple, handling the reading and writing of strings to and from the world state. All contracts for use in chaincode must implement the [contractapi.ContractInterface](https://godoc.org/github.com/hyperledger/fabric-contract-api-go/contractapi#ContractInterface). The easiest way to do this is to embed the `contractapi.Contract` struct within your own contract which will provide default functionality for meeting this interface.

Begin your contract by creating a new file `simple-contract.go` within your `contract-tutorial` folder. Within this file create a struct called `SimpleContract` which embeds the `contractapi.Contract` struct. This will be our contract for managing data to and from the world state.

```
package main

import (
    "errors"
    "fmt"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract contract for handling writing and reading from the world state
type SimpleContract struct {
    contractapi.Contract
}
```

## Writing contract functions
By default all public functions of a struct are assumed to be callable via the final chaincode; they must match a set of rules. If a public function of a contract used in a chaincode does not meet these rules then an error will be returned on chaincode creation. The rules are as follows:

- Function of contracts may only take as parameter types the following:
    - string
    - bool
    - int (including int8, int16, int32 and int64)
    - uint (including uint8, uint16, uint32 and uint64)
    - float32
    - float64
    - time.Time
    - Arrays/slices of any allowable type
    - Structs (whose public fields are all of the allowable types or another struct)
    - Pointers to structs
    - Maps with a key of type string and values of any of the allowable types
    - interface{} (Only allowed when directly taken in, will receive a string type when called via a transaction)
- Functions of contracts may also take the transaction context provided that:
    - It is taken as the first parameter
    - Either
        - It is either of type *contractapi.TransactionContext or a custom transaction context defined in the chaincode as to be used for the contract.
        - It is an interface which the transaction context type in use for the contract meets e.g. [contractapi.TransactionContextInterface](https://godoc.org/github.com/hyperledger/fabric-contract-api-go/contractapi#TransactionContextInterface)
- Functions of contracts may only return zero, one or two values
    - If the function is defined to return zero values then a success response will be returned for all calls to that contract function
    - If the function is defined to return one value then that value may be any of the allowable types listed for parameters (except `interface{}`) or `error`.
    - If the function is defined to return two values then the first may be any of the allowable types listed for parameters (except `interface{}`) and the second must be `error`

The first function to write for the simple contract is `Create`. This will add a new key value pair to the world state using a key and value provided by the user. As it interacts with the world state we will need the transaction context to be passed. We will take the default transaction context provided by contractapi (`contractapi.TransactionContext`) as it provides all the necessary functions for interacting with the world state. Taking directly `contractapi.TransactionContext` does however pose some problems, what if we were to write unit tests for our contract? We would have to create an instance of that type which would then require a [stub](https://godoc.org/github.com/hyperledger/fabric-chaincode-go/shim#ChaincodeStub) instance and would end up making our tests complex. Instead what we can do is take an interface which the transaction context meets; fortunately the contractapi package defines one: `contractapi.TransactionContextInterface`. This means if we were to write unit tests we could send a mock transaction context which could then be used to track calls or just simplify our test setup. As the function is intended to write rather than return data it will only return the error type.

```
// Create adds a new key with value to the world state
func (sc *SimpleContract) Create(ctx contractapi.TransactionContextInterface, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing != nil {
        return fmt.Errorf("Cannot create world state pair with key %s. Already exists", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}
```

The function uses the stub of the transaction context ([shim.ChaincodeStubInterface](https://godoc.org/github.com/hyperledger/fabric-chaincode-go/shim#ChaincodeStubInterface)) to first read from the world state, checking that no value exists with the supplied key, and then puts a new value into the world state, converting the passed value to a byte array as required.

The second function to add to the contract is `Update`, this will work in the same way as the Create function however instead of erroring if the key exists in the world state, it will error if it does not.

```
// Update changes the value with key in the world state
func (sc *SimpleContract) Update(ctx contractapi.TransactionContextInterface, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return fmt.Errorf("Cannot update world state pair with key %s. Does not exist", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}
```

The third and final function to add to the simple contract is `Read`. This will take in a key and return the world state value. It will therefore return a string type (the value type before converting to bytes for the world state) and will also return an error type.

```
// Read returns the value at key in the world state
func (sc *SimpleContract) Read(ctx contractapi.TransactionContextInterface, key string) (string, error) {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return "", errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

    return string(existing), nil
}
```

Your final contract will then look like this:

```
package main

import (
    "errors"
    "fmt"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract contract for handling writing and reading from the world state
type SimpleContract struct {
    contractapi.Contract
}

// Create adds a new key with value to the world state
func (sc *SimpleContract) Create(ctx contractapi.TransactionContextInterface, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing != nil {
        return fmt.Errorf("Cannot create world state pair with key %s. Already exists", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}

// Update changes the value with key in the world state
func (sc *SimpleContract) Update(ctx contractapi.TransactionContextInterface, key string, value string) error {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return fmt.Errorf("Cannot update world state pair with key %s. Does not exist", key)
    }

    err = ctx.GetStub().PutState(key, []byte(value))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}

// Read returns the value at key in the world state
func (sc *SimpleContract) Read(ctx contractapi.TransactionContextInterface, key string) (string, error) {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return "", errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

    return string(existing), nil
}
```

## Using contracts in chaincode
In the same folder as your `simple-contract.go` file, create a file called `main.go`. In here add a `main` function. This will be called when your go program is run.

```
package main

import (
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
}
```

So far you have created a contract, fabric however uses chaincode which must meet the [shim.Chaincode](https://godoc.org/github.com/hyperledger/fabric-chaincode-go/shim#Chaincode) interface. The chaincode interface requires two functions Init and Invoke. Fortunately you do not need to write these since contractapi provides a way of generating a chaincode from one or more contracts. To create a chaincode add the following to your `main` function:

```
    simpleContract := new(SimpleContract)

    cc, err := contractapi.NewChaincode(simpleContract)

    if err != nil {
        panic(err.Error())
    }
```

Once you have your chaincode, to make it callable via transactions you must start it. To do this add the following below where you create your chaincode:

```
    if err := cc.Start(); err != nil {
        panic(err.Error())
    }
```

Your `main.go` file should now look like this:

```
package main

import (
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
    simpleContract := new(SimpleContract)

    cc, err := contractapi.NewChaincode(simpleContract)

    if err != nil {
        panic(err.Error())
    }

    if err := cc.Start(); err != nil {
        panic(err.Error())
    }
}
```

## Testing your chaincode as a developer
Open a terminal to where you have cloned `fabric-samples` and cd into the `chaincode-docker-devmode` folder. This folder provides a docker-compose file defining a simple fabric network which we will run our chaincode on.

Startup the simple fabric network using:

> Note: this command will continuously print output and will not exit

```
docker-compose -f docker-compose-simple.yaml up
```

### Running the chaincode

The peer in the fabric network you have just setup is running in devmode. This means that you can start the chaincode manually yourself using the following commands.

In a new terminal window (still in the same `chaincode-docker-devmode` folder) enter the following command to enter the chaincode docker container:

```
docker exec -it chaincode sh
```

The docker-compose setup mirrors within the chaincode container the `chaincode` folder where you created your `contract-tutorial` folder. In the docker container you can therefore enter that folder by issuing the command:

```
cd contract-tutorial
```

Once in that folder you must build your chaincode program for running. You must also 'vendor' the imports for the go program as the peer will be missing these packages.

> Note: ensure you have the 2.x.x version of the fabric docker images or the `go build` command will fail.
> Note: ensure you have the correct permissions configured on your contract-tutorial folder for docker to create files there. Running `chmod -R 766` should set the correct permission levels.

```
go mod vendor
go build
```

Now run the chaincode:

> Note: it should not exit

```
CORE_CHAINCODE_ID_NAME=mycc:0 CORE_PEER_TLS_ENABLED=false ./contract-tutorial -peer.address peer:7052
```

### Interacting with the chaincode

In another new terminal window enter the CLI docker container:

```
docker exec -it cli sh
```

Despite being in devmode you still have to install the chaincode. To do this use the following command:

```
peer chaincode install -p chaincodedev/chaincode/contract-tutorial -n mycc -v 0
```

Next instantiate the chaincode so that you can start talking via the peer to it. Passing no arguments to instantiate means that no function of your contract is called.

```
peer chaincode instantiate -n mycc -v 0 -c '{"Args":[]}' -C myc
```

Once the chaincode is instantiated you can then issue transactions to call functions of your contract within the chaincode. First use an invoke to create a new key pair in the world state:

```
peer chaincode invoke -n mycc -c '{"Args":["Create", "KEY_1", "VALUE_1"]}' -C myc
```

The first argument of the invoke is the function you wish to call. This is namespaced to a contract using a colon however since you only have one contract in your chaincode you can simply pass the function name. The following arguments then make up the values that will be sent into the function. The arguments in fabric sent to a chaincode are always strings, however as described earlier a contract function can take non-string types. The contractapi generated chaincode handles conversion of these values (although in this case our function takes in strings); you can learn more about this in later tutorials. Note that you don't have to specify the transaction context despite the `Create` function taking one, this is generated for you.

Now you have created your key value pair you can use the update function of your contract to change the value. This again can be done by issuing an invoke command in the CLI container:

```
peer chaincode invoke -n mycc -c '{"Args":["Update", "KEY_1", "VALUE_2"]}' -C myc
```

You can then read the value stored for a key by issuing a query command against the read function of the contract:

```
peer chaincode query -n mycc -c '{"Args":["Read", "KEY_1"]}' -C myc
```

You should see "VALUE_2" returned.

## What to do next?
Follow the [Using advanced features](./using-advanced-features.md) tutorial.