package llm

import (
	"testing"

	wx "github.com/IBM/watsonx-go/pkg/models"
)

func TestMessageAndToolTrajectoryEvents(t *testing.T) {
	message := stringMessage(wx.RoleUser, "question")
	event := messageEvent(message)
	if event.Type != "message" || event.Role != wx.RoleUser || event.Content != "question" {
		t.Fatalf("messageEvent() = %#v", event)
	}

	tool := toolMessage("call-1", "result")
	if tool.Role != wx.RoleTool || tool.ToolCallID == nil || *tool.ToolCallID != "call-1" || tool.Content.GetText() != "result" {
		t.Fatalf("toolMessage() = %#v", tool)
	}
}

func TestStringContentPreservesText(t *testing.T) {
	content := stringContent("hello")
	if content.GetText() != "hello" {
		t.Fatalf("GetText() = %q", content.GetText())
	}
}
