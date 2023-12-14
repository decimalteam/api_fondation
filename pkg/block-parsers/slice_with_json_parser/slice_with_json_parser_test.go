package slice_with_json_parser_test

import (
	"bitbucket.org/decimalteam/api_fondation/pkg/block-parsers/slice_with_json_parser"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser/cosmos"
	"github.com/stretchr/testify/assert"
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

	parserEntity := slice_with_json_parser.NewSliceJsonParserState[cosmos.Attribute, TestTarget](testDataForParsing, "", "", "")

	parserEntity.Parse()

	result := parserEntity.Target
	assert.Equal(t, "testcoin01", result.Id)
	assert.Equal(t, "d01mx0yjpwhv0e2982rrtxhfqcpzuevz3tau78qqq", result.Delegator)
	assert.Equal(t, "d0valoper10heukkrmfyz5pktkh2pkxsh8w7q2p891111111", result.Validator)
	assert.Equal(t, "STAKE_TYPE_COIN", result.Type)
	assert.Equal(t, "testcoin01", result.Symbol)
	assert.Equal(t, "12010000000000000000", result.Amount)
	assert.Equal(t, int64(0), result.UpdatedAt)
}

func TestMapStringParseWithKind(t *testing.T) {
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
			Value: "{\"type\":\"STAKE_TYPE_COIN\",\"id\":\"testcoin01\",\"stake\":{\"denom\":\"testcoin02\",\"amount\":\"12010000000000000000\"},\"sub_token_ids\":[]}",
		},
	}

	type TestTarget struct {
		Id        string `path:"stake.stake.denom" kind:"id"`
		Delegator string `path:"delegator"`
		Type      string `path:"stake.type"`
		Symbol    string `path:"stake.stake.denom" kind:"sym"`
		UpdatedAt int64
	}

	parserEntity := slice_with_json_parser.NewSliceJsonParserState[cosmos.Attribute, TestTarget](testDataForParsing, "id", "", "")

	parserEntity.Parse()

	result := parserEntity.Target
	assert.Equal(t, "testcoin02", result.Id)
	assert.Equal(t, "d01mx0yjpwhv0e2982rrtxhfqcpzuevz3tau78qqq", result.Delegator)
	assert.Equal(t, "STAKE_TYPE_COIN", result.Type)
	assert.Equal(t, "", result.Symbol)

	parserEntity1 := slice_with_json_parser.NewSliceJsonParserState[cosmos.Attribute, TestTarget](testDataForParsing, "sym", "", "")

	parserEntity1.Parse()

	result1 := parserEntity1.Target
	assert.Equal(t, "", result1.Id)
	assert.Equal(t, "d01mx0yjpwhv0e2982rrtxhfqcpzuevz3tau78qqq", result1.Delegator)
	assert.Equal(t, "STAKE_TYPE_COIN", result1.Type)
	assert.Equal(t, "testcoin02", result1.Symbol)
}
