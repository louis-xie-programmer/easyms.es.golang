package dto

import (
	"easyms-es/easyes"
	ms "easyms-es/protos/messages"
)

// MapperToPdToken 将easyes.Tokens 转换成 ms.Tokens
func MapperToPdToken(tokens *easyes.Tokens) *ms.Tokens {
	var pdTokens ms.Tokens
	for _, token := range tokens.Tokens {
		pdToken := ms.Token{
			Token:       token.Token,
			StartOffset: token.StartOffset,
			EndOffset:   token.EndOffset,
			Type:        token.Type,
			Position:    token.Position,
			OldToken:    token.OldToken,
		}
		pdTokens.Tokens = append(pdTokens.Tokens, &pdToken)
	}
	return &pdTokens
}
