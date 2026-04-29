package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/gvamaresh/logposesift/agents/orchestrator"
	"github.com/gvamaresh/logposesift/internal/wrappers"
)

func main() {
	// Define a command-line flag named "mode". Defaults to "mcp".
	mode := flag.String("mode", "mcp", "Execution mode: 'mcp' to run the server, 'ai' to run the orchestrator")
	flag.Parse()

	// Route the execution based on the flag
	if *mode == "ai" {
		runOrchestrator()
	} else if *mode == "mcp" {
		runMCPServer()
	} else {
		fmt.Printf("Unknown mode: %s. Please use --mode=mcp or --mode=ai\n", *mode)
		os.Exit(1)
	}
}

// ---------------------------------------------------------
// AI ORCHESTRATOR LOGIC (The "Brain")
// ---------------------------------------------------------
func runOrchestrator() {
	fmt.Println("[*] Booting LogPoseSIFT AI Orchestrator...")

	// Initialize the AI Engine
	aiEngine := orchestrator.NewEngine()

	// Provide the absolute path to the extracted file on the bridge!
	evidencePath := "/mnt/sift_data/win7-32-nromanoff-memory/win7-32-nromanoff-memory-raw.001" 

	// Fire the loop!
	aiEngine.RunTriage(evidencePath)
}

// ---------------------------------------------------------
// MCP SERVER LOGIC (The "Hands")
// ---------------------------------------------------------
func runMCPServer() {
	fmt.Println("[*] Initializing LogPoseSIFT Custom MCP Server...")

	s := server.NewMCPServer(
		"LogPoseSIFT-Engine",
		"1.0.0",
		server.WithLogging(),
	)

	// Define the Volatility Tool
	windowsInfoTool := mcp.NewTool("analyze_memory_windows_info",
		mcp.WithDescription("Extracts basic OS information and kernel details from a Windows memory dump using Volatility 3."),
		mcp.WithString("dump_path", mcp.Required(), mcp.Description("Absolute file path to the memory dump.")),
	)

	// Register the Volatility Wrapper Logic
	s.AddTool(windowsInfoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		
		// 1. Tell Go to treat the raw 'any' Arguments as a Map
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments format from AI"), nil
		}

		// 2. Safely extract the dump_path string from that Map
		dumpPath, ok := args["dump_path"].(string)
		if !ok {
			return mcp.NewToolResultError("dump_path argument is missing or not a string"), nil
		}

		// 3. Execute the secure wrapper
		output, err := wrappers.GetWindowsInfo(dumpPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Tool Execution Failed: %v", err)), nil
		}

		return mcp.NewToolResultText(output), nil
	})

	// Start listening on standard I/O
	fmt.Println("[*] MCP Server is actively listening for AI commands...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}