package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	wx "github.com/IBM/watsonx-go/pkg/models"

	"github.com/ramenoodles/IBMHackathon/backend/internal/search"
	"github.com/ramenoodles/IBMHackathon/backend/internal/source"
)

const maxToolOutput = 16000

type repositoryTools struct {
	reader *source.Reader
	finder *search.Finder
	root   string
	tools  []registeredTool
}

func newRepositoryTools(reader *source.Reader, finder *search.Finder, root string) *repositoryTools {
	tools := &repositoryTools{reader: reader, finder: finder, root: root}
	tools.registerTools()
	return tools
}

type registeredTool struct {
	definition wx.ChatTool
	execute    func(context.Context, map[string]json.RawMessage) (string, error)
}

func (tools *repositoryTools) registerTools() {
	tools.registerTool("read_file", "Read a UTF-8 source file under the repository root.", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"path"},
	}, tools.readFile)
	tools.registerTool("read_context", "Read source lines around a 1-based line number.", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type": "string",
			},
			"line": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
			},
			"before": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
			},
			"after": map[string]interface{}{
				"type":    "integer",
				"minimum": 0,
			},
		},
		"required": []string{"path", "line"},
	}, tools.readContext)
	tools.registerTool("search_symbol", "Find likely function or method declarations in the repository.", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type": "string",
			},
			"language": map[string]interface{}{
				"type": "string",
			},
			"limit": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
				"maximum": 20,
			},
		},
		"required": []string{"name"},
	}, tools.searchSymbol)
}

func (tools *repositoryTools) registerTool(
	name string,
	description string,
	parameters interface{},
	handler func(context.Context, map[string]json.RawMessage) (string, error),
) {
	tools.tools = append(tools.tools, registeredTool{
		definition: wx.CreateFunction(name, description, parameters),
		execute:    handler,
	})
}

func (tools *repositoryTools) definitions() []wx.ChatTool {
	definitions := make([]wx.ChatTool, 0, len(tools.tools))
	for _, tool := range tools.tools {
		definitions = append(definitions, tool.definition)
	}
	return definitions
}

func (tools *repositoryTools) execute(ctx context.Context, call wx.ChatToolCall) (string, error) {
	var args map[string]json.RawMessage
	if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	for _, tool := range tools.tools {
		if tool.definition.Function.Name == call.Function.Name {
			return tool.execute(ctx, args)
		}
	}
	return "", fmt.Errorf("unknown tool %q", call.Function.Name)
}

func (tools *repositoryTools) readFile(_ context.Context, args map[string]json.RawMessage) (string, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := decodeArgs(args, &input); err != nil {
		return "", err
	}
	content, err := tools.reader.ReadFile(input.Path)
	if err != nil {
		return "", err
	}
	return limitOutput(fmt.Sprintf("File: %s\n\n%s", input.Path, content)), nil
}

func (tools *repositoryTools) readContext(_ context.Context, args map[string]json.RawMessage) (string, error) {
	var input struct {
		Path   string `json:"path"`
		Line   int    `json:"line"`
		Before int    `json:"before"`
		After  int    `json:"after"`
	}
	if err := decodeArgs(args, &input); err != nil {
		return "", err
	}
	result, err := tools.reader.ReadContext(input.Path, input.Line, input.Before, input.After)
	if err != nil {
		return "", err
	}
	return limitOutput(fmt.Sprintf(
		"File: %s\nLines: %d-%d\n\n%s",
		input.Path,
		result.StartLine,
		result.EndLine,
		result.Content,
	)), nil
}

func (tools *repositoryTools) searchSymbol(ctx context.Context, args map[string]json.RawMessage) (string, error) {
	var input struct {
		Name     string `json:"name"`
		Language string `json:"language"`
		Limit    int    `json:"limit"`
	}
	if err := decodeArgs(args, &input); err != nil {
		return "", err
	}
	if input.Limit <= 0 || input.Limit > 20 {
		input.Limit = 10
	}
	matches, err := tools.finder.Find(ctx, search.Query{Name: input.Name, Language: input.Language, Root: tools.root, Limit: input.Limit})
	if err != nil {
		return "", err
	}
	return limitOutput(marshalToolResult(matches)), nil
}

func decodeArgs(input map[string]json.RawMessage, output interface{}) error {
	data, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("encode arguments: %w", err)
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("decode arguments: %w", err)
	}
	return nil
}

func limitOutput(value string) string {
	if len(value) > maxToolOutput {
		return value[:maxToolOutput] + "\n[output truncated]"
	}
	return value
}

func marshalToolResult(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
	return string(data)
}
