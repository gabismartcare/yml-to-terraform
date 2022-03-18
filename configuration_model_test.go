package main

import (
	"encoding/json"
	"log"
	"testing"
)

func TestParse(t *testing.T) {
	jsonStr := `
{
  "elligibilityCriteriaAdded": {
    "clinicalTrialId": "unknow",
    "rule": {
      "type": "EXCLUSION",
      "ruleCode": "PSGDON",
      "ruleDescription": "Subject already done the PSG wearing GBB for this study",
      "bb": {
        "a": "b",
        "b": "c"
      }
    },
    "ruleId": "13f5a082-15d3-479a-8e24-01392192ccbf",
    "tableau": [
      "coucou",
      "cou",
      "cou"
    ]
  }
}
	`
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		t.Fatal(err)
	}
	toto := parse("", m)
	for k, v := range toto {
		log.Printf("%s \t : %s", k, v)
	}
}
