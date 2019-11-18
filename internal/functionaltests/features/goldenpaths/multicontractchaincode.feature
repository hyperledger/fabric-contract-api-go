# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
@goldenpaths
@multi
Feature: Multiple Contract Chaincode Golden Path

    Golden path of chaincode made up of multiple contracts

    Scenario: Initialise
        Given I have created chaincode from multiple contracts
            | SimpleContract | ComplexContract |
        Then I should be able to initialise the chaincode

    Scenario: Talk to simple contract
        When I submit the "Create" transaction
            | KEY_1 |
        And I submit the "Read" transaction
            | KEY_1 |
        Then I should receive a successful response "Initialised"

    Scenario: Talk to complex contract
        When I submit the "ComplexContract:NewObject" transaction
            | OBJECT_1 | {"name": "Andy", "contact": "Leave well alone"} | 1000 | ["red", "white", "blue"] |
        And I submit the "ComplexContract:GetObject" transaction
            | OBJECT_1 |
        Then I should receive a successful response '{"id":"OBJECT_1","owner":{"name":"Andy","contact":"Leave well alone"},"value":1000,"condition":0,"colours":["red","white","blue"]}'