package openai

import (
	"github.com/Zerofisher/goai/pkg/llm"
)

func init() {
	// Register OpenAI factory with the llm package
	llm.RegisterClientFactory("openai", func(config llm.ClientConfig) (llm.Client, error) {
		return NewClient(config)
	})
}
