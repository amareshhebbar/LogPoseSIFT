package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gvamaresh/logposesift/internal/wrappers"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/liushuangls/go-anthropic/v2"
	"google.golang.org/api/option"
)

// Engine holds BOTH clients so we can fall back seamlessly
type Engine struct {
	AnthropicClient *anthropic.Client
	AnthropicModel  anthropic.Model
	GeminiClient    *genai.Client
	GeminiModel     *genai.GenerativeModel
}

func NewEngine() *Engine {
	godotenv.Load()
	eng := &Engine{}

	// 1. Initialize Claude (if key exists)
	antKey := os.Getenv("ANTHROPIC_API_KEY")
	if antKey != "" {
		eng.AnthropicClient = anthropic.NewClient(antKey)
		eng.AnthropicModel = anthropic.Model("claude-4-6-sonnet-latest")
	}

	// 2. Initialize Gemini (if key exists)
	gemKey := os.Getenv("GEMINI_API_KEY")
	if gemKey != "" {
		client, err := genai.NewClient(context.Background(), option.WithAPIKey(gemKey))
		if err == nil {
			eng.GeminiClient = client
			// FIX: Locked to 2.5 Flash to explicitly bypass the Gemini 3.0 "Thought Signature" requirement
			eng.GeminiModel = client.GenerativeModel("gemini-2.5-flash-lite")
		} else {
			fmt.Printf("Warning: Failed to initialize Gemini client: %v\n", err)
		}
	}

	if eng.AnthropicClient == nil && eng.GeminiClient == nil {
		log.Fatal("CRITICAL: Neither ANTHROPIC_API_KEY nor GEMINI_API_KEY is set in .env")
	}

	return eng
}

func (e *Engine) TestConnection() {
	fmt.Println("[*] Dual-Engine Bridge is Active.")
}

// RunTriage acts as the Traffic Cop (Try Claude -> Fallback to Gemini)
func (e *Engine) RunTriage(evidencePath string) {
	fmt.Println("\n[*] AI Orchestrator initialized. Beginning autonomous triage...")

	if e.AnthropicClient != nil {
		fmt.Println("[*] Primary Engine Selected: Claude 4.6 Sonnet")
		err := e.runClaude(evidencePath)
		if err != nil {
			fmt.Printf("\n[!] Claude Engine Failed: %v\n", err)
			fmt.Println("[*] ===================================================")
			fmt.Println("[*] FAILOVER TRIGGERED: Rerouting request to Gemini...")
			fmt.Println("[*] ===================================================")
			
			if e.GeminiClient != nil {
				err := e.runGemini(evidencePath)
				if err != nil {
					log.Fatalf("\n[!] FATAL: Gemini Engine Failed: %v\n", err)
				}
			} else {
				log.Fatal("Gemini fallback failed: No Gemini API Key configured.")
			}
		}
	} else if e.GeminiClient != nil {
		fmt.Println("[*] Primary Engine (Claude) missing. Defaulting to Gemini...")
		err := e.runGemini(evidencePath)
		if err != nil {
			log.Fatalf("\n[!] FATAL: Gemini Engine Failed: %v\n", err)
		}
	}
}

// ---------------------------------------------------------
// CLAUDE EXECUTION LOGIC
// ---------------------------------------------------------
func (e *Engine) runClaude(evidencePath string) error {
	volatilityTool := anthropic.ToolDefinition{
		Name:        "analyze_memory_windows_info",
		Description: "Extracts basic OS info from a Windows memory dump using Volatility 3.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dump_path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute file path to the memory dump.",
				},
			},
			"required": []string{"dump_path"},
		},
	}

	messages := []anthropic.Message{
		anthropic.NewUserTextMessage(fmt.Sprintf("You are LogPoseSIFT, an autonomous DFIR agent. I have a memory dump at: %s\nPlease run the windows.info tool against it and summarize the target system's OS version.", evidencePath)),
	}

	fmt.Println("[*] Sending task to Claude...")
	resp, err := e.AnthropicClient.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model:     e.AnthropicModel,
		Messages:  messages,
		MaxTokens: 1000,
		Tools:     []anthropic.ToolDefinition{volatilityTool},
	})

	if err != nil {
		return err 
	}

	var toolUse *anthropic.MessageContentToolUse
	for _, block := range resp.Content {
		if block.Type == anthropic.MessagesContentTypeToolUse {
			toolUse = block.MessageContentToolUse
			break
		}
	}

	if toolUse == nil {
		return fmt.Errorf("Claude decided not to use any tools")
	}

	fmt.Printf("\n[*] Claude requested tool: %s\n", toolUse.Name)

	var args map[string]interface{}
	json.Unmarshal(toolUse.Input, &args)
	dumpPath := args["dump_path"].(string)

	fmt.Printf("[*] Executing Volatility on: %s\n", dumpPath)
	output, err := wrappers.GetWindowsInfo(dumpPath)
	
	toolResultContent := output
	if err != nil {
		toolResultContent = fmt.Sprintf("Error: %v", err)
	}

	messages = append(messages, anthropic.Message{
		Role:    anthropic.RoleAssistant,
		Content: resp.Content,
	})

	resultMap := map[string]interface{}{
		"type":        "tool_result",
		"tool_use_id": toolUse.ID,
		"content":     toolResultContent,
	}
	resultBytes, _ := json.Marshal(resultMap)
	var toolResultBlock anthropic.MessageContent
	json.Unmarshal(resultBytes, &toolResultBlock)

	messages = append(messages, anthropic.Message{
		Role: anthropic.RoleUser,
		Content: []anthropic.MessageContent{toolResultBlock},
	})

	finalResp, err := e.AnthropicClient.CreateMessages(context.Background(), anthropic.MessagesRequest{
		Model:     e.AnthropicModel,
		Messages:  messages,
		MaxTokens: 1000,
	})

	if err != nil {
		return err
	}

	fmt.Printf("\n================ CLAUDE'S FINAL REPORT ================\n")
	fmt.Println(finalResp.Content[0].Text)
	fmt.Printf("=======================================================\n\n")
	return nil
}

// ---------------------------------------------------------
// GEMINI EXECUTION LOGIC
// ---------------------------------------------------------
func (e *Engine) runGemini(evidencePath string) error {
	ctx := context.Background()

	volatilityTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "analyze_memory_windows_info",
			Description: "Extracts basic OS info from a Windows memory dump using Volatility 3.",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"dump_path": {
						Type:        genai.TypeString,
						Description: "Absolute file path to the memory dump.",
					},
				},
				Required: []string{"dump_path"},
			},
		}},
	}
	e.GeminiModel.Tools = []*genai.Tool{volatilityTool}

	session := e.GeminiModel.StartChat()
	prompt := fmt.Sprintf("You are LogPoseSIFT, an autonomous DFIR agent. I have a memory dump at: %s\nPlease run the analyze_memory_windows_info tool against it and summarize the target system's OS version.", evidencePath)

	fmt.Println("[*] Sending task to Gemini...")
	resp, err := session.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("Gemini API error: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return fmt.Errorf("Gemini returned an empty response (possible safety block)")
	}

	var toolCall *genai.FunctionCall
	for _, part := range resp.Candidates[0].Content.Parts {
		if fc, ok := part.(genai.FunctionCall); ok {
			toolCall = &fc
			break
		}
	}

	if toolCall == nil {
		return fmt.Errorf("Gemini decided not to use any tools")
	}

	fmt.Printf("\n[*] Gemini requested tool: %s\n", toolCall.Name)

	dumpPath := toolCall.Args["dump_path"].(string)

	fmt.Printf("[*] Executing Volatility on: %s\n", dumpPath)
	output, err := wrappers.GetWindowsInfo(dumpPath)
	
	toolResultContent := output
	if err != nil {
		toolResultContent = fmt.Sprintf("Error: %v", err)
	}

	finalResp, err := session.SendMessage(ctx, genai.FunctionResponse{
		Name: toolCall.Name,
		Response: map[string]any{
			"terminal_output": toolResultContent,
		},
	})
	
	if err != nil {
		return fmt.Errorf("Gemini final response error: %v", err)
	}

	fmt.Printf("\n================ GEMINI'S FINAL REPORT ================\n")
	for _, part := range finalResp.Candidates[0].Content.Parts {
		fmt.Println(part)
	}
	fmt.Printf("=======================================================\n\n")
	return nil
}