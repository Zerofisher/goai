package anthropic

import (
	"github.com/Zerofisher/goai/pkg/llm"
)

func init() {
	// Register Anthropic factory with the llm package
	llm.RegisterClientFactory("anthropic", func(config llm.ClientConfig) (llm.Client, error) {
		return NewClient(config)
	})
}
