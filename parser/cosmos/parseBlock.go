package cosmos

type Block struct {
	ID     interface{} `json:"id"`
	Header Header      `json:"header"`
	Data   Data        `json:"data"`
}

type Data struct {
	Time              string `json:"time"`
	Height            int    `json:"height"`
	DataTx            int    `json:"data"`
	Header            Header `json:"header"`
	Rewards           int    `json:"rewards"`
	Emission          int    `json:"emission"`
	Evidence          int    `json:"evidence"`
	LastCommit        int    `json:"last_commit"`
	StageChange       int    `json:"stage_change"`
	EndBlockEvents    int    `json:"end_block_events"`
	BeginBlockEvents  int    `json:"begin_block_events"`
	CommissionRewards int    `json:"commission_rewards"`
}

type Header struct {
	Time   string `json:"time"`
	Height int    `json:"height"`
}
