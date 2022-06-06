package hsleaderboards

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type BGResponse struct {
	Timestamp int64
	Season    int    `json:"seasonId"`
	Region    string `json:"region"`
	BGData    BGData `json:"leaderboard"`
}

type BGData struct {
	ID   string
	Rows map[string]BGRow
}

type BGRow struct {
	Name   string `json:"accountid"`
	Rank   int    `json:"rank"`
	Rating int    `json:"rating"`
}

func (receiver *BGData) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var rowsData = gjson.Get(jsonStr, "rows").String()
	var names = map[string]int{}
	receiver.ID = gjson.Get(jsonStr, "leaderboard_id").String()
	receiver.Rows = make(map[string]BGRow)
	rows := make([]BGRow, 0)
	err := json.Unmarshal([]byte(rowsData), &rows)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if val, ok := names[row.Name]; ok {
			names[row.Name] = val + 1
			row.Name = fmt.Sprintf("%s|%d", row.Name, val+1)
		} else {
			names[row.Name] = 1
		}
		receiver.Rows[row.Name] = row
	}
	return nil
}
