package hsleaderboards

import (
	"encoding/json"
	"strconv"

	"github.com/tidwall/gjson"
)

type WLResponse struct {
	Timestamp int64
	Season    int    `json:"seasonId"`
	Region    string `json:"region"`
	WLData    WLData `json:"leaderboard"`
	WLMeta    WLMeta `json:"metaData"`
}

type WLMeta struct {
	Latest int
}

type WLData struct {
	ID   string
	Rows map[string]WLRow
}

type WLRow struct {
	Name string `json:"accountid"`
	Rank int    `json:"rank"`
}

func (receiver *WLData) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var rowsData = gjson.Get(jsonStr, "rows").String()
	receiver.ID = gjson.Get(jsonStr, "leaderboard_id").String()
	receiver.Rows = make(map[string]WLRow)
	rows := make([]WLRow, 0)
	err := json.Unmarshal([]byte(rowsData), &rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		receiver.Rows[row.Name] = row
	}
	return nil
}

func (receiver *WLMeta) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var seasonsData = gjson.Get(jsonStr, "WLD.seasonsWithStartDate|@keys")
	for _, season := range seasonsData.Array() {
		val, err := strconv.Atoi(season.String())
		if err == nil && val > receiver.Latest {
			receiver.Latest = val
		}
	}
	return nil
}
