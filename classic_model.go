package hsleaderboards

import (
	"encoding/json"
	"strconv"

	"github.com/tidwall/gjson"
)

type CLResponse struct {
	Timestamp int64
	Season    int    `json:"seasonId"`
	Region    string `json:"region"`
	CLData    CLData `json:"leaderboard"`
	CLMeta    CLMeta `json:"metaData"`
}

type CLMeta struct {
	Latest int
}

type CLData struct {
	ID   string
	Rows map[string]CLRow
}

type CLRow struct {
	Name string `json:"accountid"`
	Rank int    `json:"rank"`
}

func (receiver *CLData) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var rowsData = gjson.Get(jsonStr, "rows").String()
	receiver.ID = gjson.Get(jsonStr, "leaderboard_id").String()
	receiver.Rows = make(map[string]CLRow)
	rows := make([]CLRow, 0)
	err := json.Unmarshal([]byte(rowsData), &rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		receiver.Rows[row.Name] = row
	}
	return nil
}

func (receiver *CLMeta) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var seasonsData = gjson.Get(jsonStr, "CLS.seasonsWithStartDate|@keys")
	for _, season := range seasonsData.Array() {
		val, err := strconv.Atoi(season.String())
		if err == nil && val > receiver.Latest {
			receiver.Latest = val
		}
	}
	return nil
}
