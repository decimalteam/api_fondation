package genesis_parser

import (
	"bitbucket.org/decimalteam/api_fondation/client"
	"bitbucket.org/decimalteam/api_fondation/pkg/parser"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"reflect"
)

// CoinScheme is for parsing genesis gained data
type CoinScheme struct {
	Crr         float64 `json:"crr"`
	Creator     string  `json:"creator"`
	Identity    string  `json:"identity"`
	LimitVolume string  `json:"limit_volume"`
	Reserve     string  `json:"reserve"`
	Denom       string  `json:"denom"`
	Title       string  `json:"title"`
	Volume      string  `json:"volume"`
}

type StateGen struct {
	CoinsData []CoinScheme
	parser    *parser.Parser
	logger    *logrus.Logger
}

type GenParser interface {
	ParseGenesis()
}

func NewDecimalParserGen(newParser *parser.Parser, logger *logrus.Logger) *StateGen {
	return &StateGen{
		CoinsData: []CoinScheme{},
		parser:    newParser,
		logger:    logger,
	}
}

func (s *StateGen) ParseGenesis() {
	apiClient, _ := client.New()
	ctx := context.Background()

	var jsonObtained string
	for i := 0; ; i++ {
		chunked, errGen := apiClient.RpcClient.GenesisChunked(ctx, uint(i))
		if errGen != nil {
			value := reflect.ValueOf(errGen)
			code := value.Elem().FieldByName("Code").Int()
			if code == -32603 {
				break
			} else {
				s.logger.Errorf("chunk getting error: %v", errGen)
			}
		}

		jsonPart, errBase := base64.StdEncoding.DecodeString(chunked.Data)
		if errBase != nil {
			s.logger.Fatalf("decode chunk error: %v", errBase)
		}

		jsonObtained = fmt.Sprint(jsonObtained, string(jsonPart))
	}

	// getting coins array from genesis chunk
	coinsArr := getValueInJsonByPath("app_state.coin.coins", []byte(jsonObtained))

	var coins []CoinScheme
	if errMarsh := json.Unmarshal([]byte(coinsArr.Raw), &coins); errMarsh != nil {
		s.logger.Fatalf("unmarshal josn error: %v", errMarsh)
	}
	s.CoinsData = coins
}

func getValueInJsonByPath(path string, obtainedJson []byte) gjson.Result {
	var jsonStruct map[string]interface{}
	jsonParseErr := json.Unmarshal(obtainedJson, &jsonStruct)
	if jsonParseErr == nil {
		return gjson.Get(string(obtainedJson), path)
	}
	return gjson.Result{}
}
