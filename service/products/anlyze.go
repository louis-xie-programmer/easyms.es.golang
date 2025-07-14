package products

import (
	"easyms-es/easyes"
)

func Analyze(analyzer string, text string) (*easyes.Tokens, error) {
	res, err := ProductStore.Analyze(analyzer, text)
	if err != nil {
		return nil, err
	}

	return res, nil
}
