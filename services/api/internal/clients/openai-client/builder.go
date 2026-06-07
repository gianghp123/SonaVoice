package openaiclient

import (
	"context"
	"errors"
	"reflect"
)

var (
	ErrPromptRequired            = errors.New("prompt is required")
	ErrStructuredOutputTargetNil = errors.New("structured output target is required")
	ErrStructuredOutputTargetPtr = errors.New("structured output target must be a pointer")
	ErrStructuredSchemaRequired  = errors.New("structured output schema is required")
)

func isPointer(v any) bool {
	return v != nil && reflect.TypeOf(v).Kind() == reflect.Ptr
}

type ResponseBuilder struct {
	ctx      context.Context
	executor IOpenAIClient

	prompt string

	withStructuredOutput bool
	structuredOutput     any
	schemaName           string
	schema               map[string]any
}

func NewResponse(ctx context.Context, executor IOpenAIClient) ResponseRequest {
	return &ResponseBuilder{
		ctx:      ctx,
		executor: executor,
	}
}

func (b *ResponseBuilder) Prompt(prompt string) ResponseRequest {
	b.prompt = prompt
	return b
}

func (b *ResponseBuilder) WithStructuredOutput(output any, schemaName string, schema map[string]any) ResponseRequest {
	b.withStructuredOutput = true
	b.structuredOutput = output
	b.schemaName = schemaName
	b.schema = schema
	return b
}

func (b *ResponseBuilder) Do() (string, error) {
	if b.prompt == "" {
		return "", ErrPromptRequired
	}

	if !b.withStructuredOutput {
		return b.executor.ExecuteText(b.ctx, b.prompt)
	}

	if b.structuredOutput == nil {
		return "", ErrStructuredOutputTargetNil
	}

	if isPointer(b.structuredOutput) {
		return "", ErrStructuredOutputTargetPtr
	}

	if b.schemaName == "" || b.schema == nil {
		return "", ErrStructuredSchemaRequired
	}

	err := b.executor.ExecuteStructured(
		b.ctx,
		b.prompt,
		b.structuredOutput,
		b.schemaName,
		b.schema,
	)
	if err != nil {
		return "", err
	}

	return "", nil
}
