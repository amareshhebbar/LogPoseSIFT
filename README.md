
<div align="center">
  <h1>LogPoseSIFT</h1>
  <p>
    <strong>Autonomous Multi-Agent DFIR Orchestrator</strong><br />
    Protecting evidence integrity while accelerating incident triage through type-safe Custom MCP Servers.<br /><br />
    <a href="https://github.com/gvamaresh/logposesift/tree/main/docs">Explore Architecture</a> ·
    <a href="https://github.com/gvamaresh/logposesift/issues">Report an Issue</a>
  </p>

  <p>
    <a href="https://github.com/gvamaresh/logposesift/commits/main">
      <img src="https://img.shields.io/github/last-commit/gvamaresh/logposesift?style=flat-square&logo=git&color=005571" alt="Last Commit" />
    </a>
    <a href="https://github.com/gvamaresh/logposesift">
      <img src="https://img.shields.io/github/repo-size/gvamaresh/logposesift?style=flat-square&logo=github&color=005571" alt="Repo Size" />
    </a>
    <a href="https://github.com/gvamaresh/logposesift/blob/main/LICENSE">
      <img src="https://img.shields.io/github/license/gvamaresh/logposesift?style=flat-square&logo=open-source-initiative&color=005571" alt="License" />
    </a>
    <a href="https://github.com/gvamaresh/logposesift">
      <img src="https://img.shields.io/badge/Language-Go-005571?style=flat-square&logo=go" alt="Built with Go" />
    </a>
  </p>
</div>

<br />

---

## Features

LogPoseSIFT equips incident responders with an AI-driven triage system that operates at machine speed without compromising forensic soundness. Here is what you can do:

* **Type-Safe Tool Execution:** Expose raw SIFT command-line tools (like Volatility and Plaso) as strictly typed Golang API endpoints to physically prevent evidence spoliation and destructive commands.
* **Multi-Agent Orchestration:** Decompose complex forensic workloads by routing tasks between specialized AI agents (e.g., Memory Specialist, Disk Specialist, Synthesizer) to prevent context window degradation.
* **Self-Correcting Execution Loops:** Utilize bounded iteration caps that allow agents to catch errors, adjust parameters, and retry tool execution without falling into infinite conversational spirals.
* **Structured Execution Logging:** Generate transparent, programmatically structured logs detailing full tool execution sequences, agent-to-agent communication, and token usage for judge review and auditability.
* **Context Window Optimization:** Parse massive terminal outputs into clean, concise JSON structures before returning data to the LLM, ensuring the system processes only highly relevant artifacts.

**Why It Matters:** These features allow security teams and practitioners to match the velocity of autonomous threat actors, reducing initial triage time from hours to seconds while maintaining absolute evidence integrity.

---

## Installation

LogPoseSIFT is designed to run natively within the SANS SIFT Workstation environment. 

### Prerequisites
* SANS SIFT Workstation VM (Ubuntu 22.04 base)
* Golang 1.21+ installed
* Protocol SIFT baseline installed

To get started, clone the repository into your SIFT workspace:

```bash
git clone [https://github.com/gvamaresh/logposesift.git](https://github.com/gvamaresh/logposesift.git)
cd logposesift
```

---

## Quick Start

Initialize the Go modules and build the custom MCP server:

```bash
go mod tidy
go build -o logposesift-mcp cmd/sift-mcp/main.go
```

### Execution

Provide your API credentials and execute the orchestrator against a target evidence file:

```bash
export API_KEY="your_api_key_here"
./logposesift-mcp --target /path/to/evidence.raw --mode triage
```

**Expected Startup Output:**

```text
2026-04-28 15:42:10.157 | INFO | logposesift.server:main - Starting LogPoseSIFT Custom MCP Server
2026-04-28 15:42:10.160 | INFO | logposesift.wrappers:load - Initializing typed wrappers for Volatility3, Plaso
2026-04-28 15:42:10.165 | INFO | logposesift.agents:orchestrator - Agent routing logic loaded
INFO: Waiting for application startup.
INFO: Server running on [http://127.0.0.1:8718](http://127.0.0.1:8718) (Press CTRL+C to quit)
```

---

## Architecture details

LogPoseSIFT utilizes a hybrid architecture:

1.  **Custom MCP Server (Golang):** Acts as the security boundary. It wraps the 200+ raw SIFT tools. The LLM cannot execute raw shell commands; it can only call predefined functions like `analyze_memory()` or `extract_timeline()`.
2.  **Agent Logic (Python/Go):** Manages the state and routing. A synthesizer agent evaluates the overall case and dispatches targeted requests to domain-specific agents, enforcing strict `--max-iterations` to prevent runaway token usage.

---

## Roadmap and Future Goals

* **Live Endpoint Integration:** Expanding the MCP server to pull live data from SIEMs or remote endpoints.
* **Expanded Tool Wrappers:** Adding strict JSON parsing for the remaining 150+ SIFT command-line utilities.
* **Benchmarking Harness:** Releasing an automated accuracy framework to measure false positives against known-good disk images.

---

## Contributing

Contributions to LogPoseSIFT are welcome. Please read the contributing guidelines before submitting a pull request.
