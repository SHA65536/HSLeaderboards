package hsleaderboards

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

type STResponse struct {
	Timestamp int64
	Season    int    `json:"seasonId"`
	Region    string `json:"region"`
	STData    STData `json:"leaderboard"`
	STMeta    STMeta `json:"metaData"`
}

type STMeta struct {
	Latest int
}

type STData struct {
	ID   string
	Rows map[string]STRow
}

type STRow struct {
	Name string `json:"accountid"`
	Rank int    `json:"rank"`
}

func (receiver *STData) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var rowsData = gjson.Get(jsonStr, "rows").String()
	var names = map[string]int{}
	receiver.ID = gjson.Get(jsonStr, "leaderboard_id").String()
	receiver.Rows = make(map[string]STRow)
	rows := make([]STRow, 0)
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

func (receiver *STMeta) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var seasonsData = gjson.Get(jsonStr, "STD.seasonsWithStartDate|@keys")
	for _, season := range seasonsData.Array() {
		val, err := strconv.Atoi(season.String())
		if err == nil && val > receiver.Latest {
			receiver.Latest = val
		}
	}
	return nil
}
