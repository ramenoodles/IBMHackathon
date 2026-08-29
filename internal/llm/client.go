package llm

import "context"

type Client interface {
	Ask(
		ctx context.Context,
		question string,
		source string,
	) (string, error)
}
