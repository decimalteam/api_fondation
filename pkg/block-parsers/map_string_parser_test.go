package block_parsers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testTarget struct {
	Symbol    string `path:"coin.denom,sf3.denom.sym"`
	Amount    string `path:"coin.amount"`
	SubField1 string `path:"coin.subfield.subsub1"`
	SubField2 string `path:"coin.subfield.subsub2"`
	Delegator string `path:"delegator"`
	Validator string `path:"validator"`
	UpdatedAt int64
}

func TestParseFrom(t *testing.T) {
	testDataForParsing := map[string]interface{}{
		"coin": map[string]interface{}{
			"symbol": "testcoin01",
			"amount": "123000000000000000", // kind `delegated`
			"subfield": map[string]interface{}{
				"subsub1": "subsubvalue1",
				"subsub2": "subsubvalue2",
			},
		},
		"delegator": "d01mx0yjpwhv0e2982rrtxhfqcpzuevz3tau78ysg",
		"validator": "d0valoper1qkatcxttjmpz7a3l0cc2snn5vedfyrucgdca4h",
		"sf3": map[string]interface{}{
			"denom": map[string]interface{}{
				"sym": "testcoin02",
			},
		},
	}

	parserEntity := NewMapStringParser(testDataForParsing, new(testTarget), "")
	parserEntity.Parse()

	result := parserEntity.Target
	assert.Equal(t, "testcoin02", result.Symbol)
	assert.Equal(t, "123000000000000000", result.Amount)
	assert.Equal(t, "subsubvalue1", result.SubField1)
	assert.Equal(t, "subsubvalue2", result.SubField2)
	assert.Equal(t, "d01mx0yjpwhv0e2982rrtxhfqcpzuevz3tau78ysg", result.Delegator)
	assert.Equal(t, "d0valoper1qkatcxttjmpz7a3l0cc2snn5vedfyrucgdca4h", result.Validator)
	assert.Equal(t, int64(0), result.UpdatedAt)
}
