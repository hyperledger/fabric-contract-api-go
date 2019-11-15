# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
@errors
@badpaths
Feature: Error paths

   Check how errors are handled by api

    Scenario: User calls unknown function when contract uses unknown transaction handler
        Given I have created and initialised chaincode "SimpleContract"
        When I submit the "FakeFunction" transaction
            | Some | Args |
        Then I should receive an unsuccessful response "Function FakeFunction not found in contract SimpleContract"

    Scenario: User calls unknown function when contract has set an unknown transaction handler
        Given I have created and initialised chaincode "ExtendedSimpleContract"
        When I submit the "FakeFunction" transaction
            | Some | Args |
        Then I should receive an unsuccessful response "Invalid function FakeFunction passed with args [Some, Args]"

    Scenario: Contract function returns error
        Given I have created and initialised chaincode "SimpleContract"
        When I submit the "Read" transaction
            | MISSING_KEY |
        Then I should receive an unsuccessful response "Cannot read key. Key with id MISSING_KEY does not exist"

    Scenario: User sends bad basic data type
        Given I have created and initialised chaincode "ComplexContract"
        When I submit the "NewObject" transaction
            | OBJECT_1 | {"name": "Andy", "contact": "Leave well alone"} | -10 | ["red", "white", "blue"] |
        Then I should receive an unsuccessful response "Error managing parameter param2. Conversion error. Cannot convert passed value -10 to uint"

    Scenario: Users sends bad object data type
        Given I have created and initialised chaincode "ComplexContract"
        When I submit the "NewObject" transaction
            | OBJECT_1 | {"firstname": "Andy", "contact": "Leave well alone"} | 1000 | ["red", "white", "blue"] |
        Then I should receive an unsuccessful response "Error managing parameter param1. Value did not match schema:\n1. param1: Additional property firstname is not allowed\n2. param1: name is required"

    Scenario: User sends data that does not match custom metadata
        Given I am using metadata file "contracts/complexcontract/contract-metadata/metadata.json"
        And I have created chaincode from "ComplexContract"
        When I submit the "NewObject" transaction
            | OBJECT_A | {"name": "Andy", "contact": "Leave well alone"} | 1000 | ["red", "white", "blue"] |
        Then I should receive an unsuccessful response "Error managing parameter param0. Value did not match schema:\n1. param0: Does not match pattern '^OBJECT_\d$'"

    Scenario: Contract returns data that does not match custom metadata
        Given I am using metadata file "contracts/complexcontract/contract-metadata/metadata.json"
        And I have created chaincode from "ComplexContract"
        And I submit the "NewObject" transaction
            | OBJECT_1 | {"name": "Andy", "contact": "Leave well alone"} | 1000 | ["red", "white", "blue"] |
        And I receive a successful response
        When I submit the "GetValue" transaction
            | OBJECT_1 |
        Then I should receive an unsuccessful response "Error handling success response. Value did not match schema:\n1. return: Must be less than or equal to 10"


    Scenario: User configures bad metadata file
        Given I am using metadata file "utils/bad_metadata.json"
        Then I fail to create chaincode from "SimpleContract" 

    Scenario: User sends invalid namespace to multi contract
        Given I have created chaincode from multiple contracts
            | SimpleContract | ComplexContract |
        When I submit the "FakeContract:NewObject" transaction
            | SomeValue |
        Then I should receive an unsuccessful response "Contract not found with name FakeContract"