package openaiclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

type OpenAIClient struct {
	sdk   *openai.Client
	model string
}

func NewOpenAIClient(sdk *openai.Client, model string) *OpenAIClient {
	return &OpenAIClient{
		sdk:   sdk,
		model: model,
	}
}

func (c *OpenAIClient) ExecuteText(ctx context.Context, prompt string) (string, error) {
	resp, err := c.sdk.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: openai.ChatModel(c.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		},
		option.WithJSONSet("thinking.type", "disabled"),
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) ExecuteStructured(
	ctx context.Context,
	prompt string,
	output any,
	schemaName string,
	schema map[string]any,
) error {
	resp, err := c.sdk.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: openai.ChatModel(c.model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
					Type: "json_schema",
					JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
						Name:   schemaName,
						Schema: schema,
						Strict: openai.Bool(true),
					},
				},
			},
		},
		option.WithJSONSet("thinking.type", "disabled"),
	)
	if err != nil {
		return err
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("no completion choices returned")
	}

	content := resp.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(content), output); err != nil {
		return fmt.Errorf("failed to unmarshal structured output: %w; raw content: %s", err, content)
	}

	return nil
}
