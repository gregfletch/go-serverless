[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Go Serverless

This is a sample serverless app using Golang. In this README, you will find instructions to setup a Go development
environment, install required dependencies, run the tests, and deploy the Serverless API app to AWS.

- [Setup](#setup)
  - [Go](#go)
  - [Node](#node)
  - [Yarn](#yarn)
- [Installing Dependencies](#installing-dependencies)
  - [Required Go Dependencies](#required-go-dependencies)
  - [Serverless](#serverless)

## Setup

### Go

Download and install the latest version of Go from following the instructions [here](https://go.dev/doc/install).

### Node

This project uses a `.nvmrc` file to automatically set the local Node version. If you are not using NVM to manage
node versions locally, you will need to download and use Node version 16.13.0 or newer.

### Yarn

Download and install the latest version of Yarn from following the instructions
[here](https://classic.yarnpkg.com/lang/en/docs/install).

## Installing Dependencies

### Required Go Dependencies

To download and install the required Go dependencies, run the following command from a terminal within the root
directory of the project:

```shell
$ go mod tidy
```

### Serverless

To download and install serverless and the local/offline serverless plugins, run the following command from a terminal
within the root directory of the project:

```shell
$ yarn install
```
