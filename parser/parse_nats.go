package parser

import (
	"bitbucket.org/decimalteam/api_fondation/nats"
	"bitbucket.org/decimalteam/api_fondation/parser/cosmos"
)

func NatsParser(ch chan cosmos.Block) {
	nats.Connect()
}
