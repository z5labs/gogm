#!/bin/bash

# run this file if changing interface definitions for ISession, ITransaction, SessionV2 or TransactionV2

go get github.com/vektra/mockery/v2/.../

mockery --name ISession
mockery --name ITransaction
mockery --name SessionV2
mockery --name TransactionV2
