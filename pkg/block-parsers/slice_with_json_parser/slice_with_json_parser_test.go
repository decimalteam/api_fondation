package slice_with_json_parser_test

import (
	"bitbucket.org/decimalteam/api_fondation/pkg/block-parsers/slice_with_json_parser"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"fmt"
	"testing"
)

func TestMapStringParse(t *testing.T) {
	testDataForParsing := []cosmos.Attribute{
		{
			Key:   "delegator",
			Value: "d01mx0yjpwhv0e2982rrtxhfqcpzuevz3tau78qqq",
		},
		{
			Key:   "validator",
			Value: "d0valoper10heukkrmfyz5pktkh2pkxsh8w7q2p891111111",
		},
		{
			Key:   "stake",
			Value: "{\"type\":\"STAKE_TYPE_COIN\",\"id\":\"testcoin01\",\"stake\":{\"denom\":\"testcoin01\",\"amount\":\"12010000000000000000\"},\"sub_token_ids\":[]}",
		},
	}

	type TestTarget struct {
		Id        string `path:"stake.id"`
		Delegator string `path:"delegator"`
		Validator string `path:"validator"`
		Type      string `path:"stake.type"`
		Symbol    string `path:"stake.stake.denom"`
		Amount    string `path:"stake.stake.amount"`
		UpdatedAt int64
	}

	parserInstance := slice_with_json_parser.NewSliceJsonParserState[cosmos.Attribute, TestTarget](testDataForParsing, "Key", "Value", "")

	parserInstance.ParseCoinDataFromAttributes()
	fmt.Println(parserInstance.Target)
}
