# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
@goldenpaths
@complex
Feature: Complex Chaincode Golden Path

    Golden path of a chaincode which handles putting and getting an object

    Scenario: Initialise
        Given I have created chaincode from "ComplexContract"
        Then I should be able to initialise the chaincode

    Scenario: Create complex object
        When I submit the "NewObject" transaction
            | OBJECT_1 | {"name": "Andy", "contact": "Leave well alone"} | 1000 | ["red", "white", "blue"] |
        Then I should receive a successful response

    Scenario: Read new complex object
        When I submit the "GetObject" transaction
            | OBJECT_1 |
        Then I should receive a successful response '{"id":"OBJECT_1","owner":{"name":"Andy","contact":"Leave well alone"},"value":1000,"condition":0,"colours":["red","white","blue"]}'

    Scenario: Update complex object owner
        When I submit the "UpdateOwner" transaction
            | OBJECT_1 | {"name": "Liam", "contact": "Bug whenever"} |
        Then I should receive a successful response
    
    Scenario: Read owner updated complex object
        When I submit the "GetObject" transaction
            | OBJECT_1 |
        Then I should receive a successful response '{"id":"OBJECT_1","owner":{"name":"Liam","contact":"Bug whenever"},"value":1000,"condition":1,"colours":["red","white","blue"]}'

    Scenario: Update complex object value
        When I submit the "UpdateValue" transaction
            | OBJECT_1 | -50 |
        Then I should receive a successful response
    
    Scenario: Read complex object
        When I submit the "GetObject" transaction
            | OBJECT_1 |
        Then I should receive a successful response '{"id":"OBJECT_1","owner":{"name":"Liam","contact":"Bug whenever"},"value":950,"condition":1,"colours":["red","white","blue"]}'
