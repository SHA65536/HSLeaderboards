package hsleaderboards

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

type MRResponse struct {
	Timestamp int64
	Season    int    `json:"seasonId"`
	Region    string `json:"region"`
	MRData    MRData `json:"leaderboard"`
	MRMeta    MRMeta `json:"metaData"`
}

type MRMeta struct {
	Latest int
}

type MRData struct {
	ID   string
	Rows map[string]MRRow
}

type MRRow struct {
	Name   string `json:"accountid"`
	Rank   int    `json:"rank"`
	Rating int    `json:"rating"`
}

func (receiver *MRData) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var rowsData = gjson.Get(jsonStr, "rows").String()
	var names = map[string]int{}
	receiver.ID = gjson.Get(jsonStr, "leaderboard_id").String()
	receiver.Rows = make(map[string]MRRow)
	rows := make([]MRRow, 0)
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

func (receiver *MRMeta) UnmarshalJSON(data []byte) error {
	var jsonStr = string(data)
	var seasonsData = gjson.Get(jsonStr, "MRC.seasonsWithStartDate|@keys")
	for _, season := range seasonsData.Array() {
		val, err := strconv.Atoi(season.String())
		if err == nil && val > receiver.Latest {
			receiver.Latest = val
		}
	}
	return nil
}
