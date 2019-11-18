# Copyright the Hyperledger Fabric contributors. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
@goldenpaths
@extended
Feature: Extended Simple Chancode Golden Path

   Golden path of a very basic put and get chaincode which uses a before transaction

   Scenario: Initialise
      Given I have created chaincode from "ExtendedSimpleContract"
      Then I should be able to initialise the chaincode

   Scenario: Create key value pair
      When I submit the "Create" transaction
         | KEY_1 |
      Then I should receive a successful response

   Scenario: Read key value pair
      When I submit the "Read" transaction
         | KEY_1 |
      Then I should receive a successful response "Initialised"

   Scenario: Update key value pair
      When I submit the "Update" transaction
         | KEY_1 | Updated |
      Then I should receive a successful response

   Scenario: Read updated key value pair
      When I submit the "Read" transaction
         | KEY_1 |
      Then I should receive a successful response "Updated"
