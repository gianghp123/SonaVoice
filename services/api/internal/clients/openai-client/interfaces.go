package openaiclient

import "context"

type IOpenAIClient interface {
	ExecuteText(ctx context.Context, prompt string) (string, error)
	ExecuteStructured(ctx context.Context, prompt string, output any, schemaName string, schema map[string]any) error
}

type ResponseRequest interface {
	Prompt(prompt string) ResponseRequest
	WithStructuredOutput(output any, schemaName string, schema map[string]any) ResponseRequest
	Do() (string, error)
}
