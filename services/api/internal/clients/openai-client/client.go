package openaiclient

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
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
	resp, err := c.sdk.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
		Model: c.model,
	})

	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

func (c *OpenAIClient) ExecuteStructured(
	ctx context.Context,
	prompt string,
	output any,
	schemaName string,
	schema map[string]any,
) error {
	response, err := c.sdk.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
		Text: responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigParamOfJSONSchema(
				schemaName,
				schema,
			),
		},
	})
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(response.OutputText()), output)
}
