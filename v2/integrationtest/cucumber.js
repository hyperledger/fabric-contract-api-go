// Copyright the Hyperledger Fabric contributors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0


// Configuration for running the chaincode tests
// eslint-disable-next-line @typescript-eslint/no-var-requires
const path = require('path');

const TEST_NETWORK_DIR= path.resolve(process.env['TEST_NETWORK_DIR'])

// it's important to note the langauge of the chaincode implementations.
const config = {
  TestNetwork: {
    rootDir: TEST_NETWORK_DIR,
    chaincodes: {
      simple: { path: path.resolve('./chaincode/simple'), lang: "golang" },
      ccaas: { path: path.resolve('./chaincode/simple'), lang: "ccaas" },
      advancedtypes: { path: path.resolve('./chaincode/advancedtypes'), lang: "golang"}
    },
    cryptoPath : path.resolve(TEST_NETWORK_DIR, 'organizations', 'peerOrganizations', 'org1.example.com'),
    env: "", 
    peerEndpoint : 'localhost:7051',
    useExisting: false
  }
}

// -----------------------------------------

// These configurations affect how the cucumber framework is used. In general these do not 
// need to be modified

// this configuration is only used when developing changes to the tool itself
let dev = [
  './features/**/*.feature', // Specify our feature files
  '--require-module ts-node/register', // Load TypeScript module
  `--require ./src/step-definitions/**/*.ts`, // Load step definitions
  '--format progress-bar', // Load custom formatter
  '--format @cucumber/pretty-formatter', // Load custom formatter,
  `--world-parameters ${JSON.stringify(config)}`,
  '--publish-quiet'
].join(' ');

// This should be in used in all other circumstances
const installDir = path.resolve(process.cwd(),'node_modules','@hyperledger', 'fabric-chaincode-integration');

let prod = [
  `${installDir}/features/**/*.feature`, // Specify our feature files
  '--require-module ts-node/register', // Load TypeScript module
  `--require ${installDir}/dist/step-definitions/**/*.js`, // Load step definitions
  '--format progress-bar', // Load custom formatter
  '--format @cucumber/pretty-formatter', // Load custom formatter,
  `--world-parameters ${JSON.stringify(config)}`,
  '--publish-quiet'
].join(' ');

module.exports = {
  default: prod,
  prod,
  dev
};

