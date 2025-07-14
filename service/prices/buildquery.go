package prices

import (
	"fmt"
	"strings"
)

func BuildSingleQuery(pid int, from int, size int) string {
	orderScript = strings.ReplaceAll(orderScript, "\r", "")
	orderScript = strings.ReplaceAll(orderScript, "\n", "")
	orderScript = strings.ReplaceAll(orderScript, "\t", "")
	return fmt.Sprintf(SingleQueryStr, pid, orderScript, from, size)
}

var orderScript = `long currentTime = new Date().getTime();
            long eventTime = doc['UpdateTime'].value.getMillis();
            long diff = currentTime - eventTime;
            long diffDays = diff / (1000 * 60 * 60 * 24);
            if (diffDays <= 1) {
              return 10 + doc['Sort'].value;
            } else if (diffDays <= 7) {
              return 7 + doc['Sort'].value;
            } else if (diffDays <= 14) {
              return 5 + doc['Sort'].value;
            } else if (diffDays <= 30) {
              return 3 + doc['Sort'].value;
            } else {
              return 1 + doc['Sort'].value;
            }`

const SingleQueryStr = `{
"query": {
    "bool": {
      "filter": [
        {
          "term": {
            "PID": "%d"
          }
        }
      ]
    }
  },
  "sort": {
    "_script":{
      "type": "number",
      "script":{
        "lang": "painless",
        "source": "%s",
        "params":{
          "factor":1.1
        }
      },
      "order":"desc"
    }
  },
  "from": %d,
  "size": %d
}`
