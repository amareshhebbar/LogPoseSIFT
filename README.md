<div align="center">
  <h1>LogPoseSIFT</h1>
  <p>
    <strong>Autonomous Multi-Agent DFIR Orchestrator</strong><br />
    100% precision · 0 hallucinations · Custom Go MCP Server · Real APT dataset validated
  </p>
  <p>
    <a href="https://github.com/amareshhebbar/LogPoseSIFT/blob/main/docs/architecture.md">Architecture</a> ·
    <a href="https://github.com/amareshhebbar/LogPoseSIFT/blob/main/docs/accuracy_report.md">Accuracy Report</a> ·
    <a href="https://github.com/amareshhebbar/LogPoseSIFT/blob/main/docs/dataset.md">Dataset</a> ·
    <a href="https://github.com/amareshhebbar/LogPoseSIFT/issues">Issues</a>
  </p>
  <p>
    <img src="https://img.shields.io/github/last-commit/amareshhebbar/LogPoseSIFT?style=flat-square&color=005571" />
    <img src="https://img.shields.io/badge/Language-Go-005571?style=flat-square&logo=go" />
    <img src="https://img.shields.io/badge/Precision-100%25-005571?style=flat-square" />
    <img src="https://img.shields.io/badge/Recall-92.8%25-005571?style=flat-square" />
    <img src="https://img.shields.io/badge/Hallucinations-0-005571?style=flat-square" />
    <img src="https://img.shields.io/github/license/amareshhebbar/LogPoseSIFT?style=flat-square&color=005571" />
  </p>
</div>


> **AI Engine:** Designed for Claude Sonnet 4.6 (primary) with Gemini 2.5 Flash as automatic failover. All benchmark results and demo video were produced using Claude. Gemini fallback is functional but untested against the full SRL-2018 dataset.

---

## What Is This

LogPoseSIFT is an autonomous DFIR triage agent that connects Claude (with Gemini failover) to thze SANS SIFT Workstation toolchain through a **Custom MCP Server written in Go**.

The core design principle: **the LLM cannot run shell commands**. It can only call typed Go functions. The MCP server is the security boundary — architectural enforcement, not prompt-based rules.

On the SRL-2018 APT dataset (a documented real-world intrusion with a DKOM rootkit, C2 beaconing, and lateral movement), LogPoseSIFT:

- Identified **13 of 14 documented IOCs** including the external C2 IP, all malicious processes, and the rootkit itself
- Achieved **100% precision** — zero hallucinated findings
- Ran fully autonomously across **6 agentic iterations** with a self-correction sequence that detected DKOM rootkit hiding

---

## Evidence & Results

| Document | Description |
|---|---|
| [Screenshots →](docs/SCREENSHOTS.md) | 5 annotated screenshots covering security boundary, real APT findings, self-correction, benchmark, and audit trail |
| [Demo Video →](docs/VIDEO.md) | 5-minute screencast with narration — live triage on SRL-2018 APT evidence |
| [Live Results →](docs/RESULT.md) | Full output from actual triage run — process findings, C2 connections, rootkit detection |

## Quickstart (SIFT Workstation)

```bash
# 1. Clone
git clone https://github.com/amareshhebbar/LogPoseSIFT
cd LogPoseSIFT

# 2. Set API key
echo "ANTHROPIC_API_KEY=sk-ant-your-key-here" > .env

# 3. Build
go mod tidy
go build -o logpose-ai ./cmd/sift-mcp/

# 4. Run MCP server (for use with Claude Desktop or Claude Code)
./logpose-ai --mode=mcp

# 5. Run autonomous triage on evidence
./logpose-ai --mode=ai \
  --target=/path/to/evidence.img \
  --type=memory        # or: disk | both
```

### Run Benchmark

```bash
# Extract SRL-2018 evidence first
7z x /path/to/base-hunt-memory.7z -o/tmp/evidence/

# Run benchmark (scores against documented ground truth)
chmod +x benchmark/run_benchmark.sh
./benchmark/run_benchmark.sh /tmp/evidence/base-hunt-memory.img memory
```

Expected output:
```
True Positives  (TP): 13
False Negatives (FN): 1
False Positives (FP): 0
Precision: 100.00%
Recall:    92.86%
```

---

## Architecture

```
Claude/Gemini (LLM)
       │ MCP calls only — no shell access
       ▼
cmd/sift-mcp/main.go  ← SECURITY BOUNDARY
       │ 12 typed MCP tools registered
       │
  ┌────┴────┐
  │         │
agents/   internal/
  │         │
  ├─ orchestrator    ├─ wrappers (7 typed tool wrappers)
  ├─ memory_agent    ├─ validator (CONFIRMED/INFERRED/UNVERIFIED)
  ├─ disk_agent      ├─ correlator (disk vs memory cross-ref)
  └─ reasoning_logger└─ registry (tool allowlist, 30+ entries)
       │
  SIFT Tools (read-only)
  vol / fls / log2timeline / rip.pl / yara / hashdeep
```

The LLM calls MCP tools → Go dispatches to typed wrappers → wrappers call SIFT binaries via `exec.Command` (never `bash -c`) → output parsed to JSON → structured result returned to LLM.

[Full architecture documentation →](docs/architecture.md)

---

## MCP Tools (12 registered)

| Tool | Category | What It Does |
|---|---|---|
| `analyze_memory_windows_info` | Memory | OS version, kernel base, architecture |
| `analyze_memory_pslist` | Memory | Process scan via pool tags (bypasses DKOM) |
| `analyze_memory_netscan` | Memory | Active + closed TCP/UDP connections |
| `analyze_memory_malfind` | Memory | Code injection, process hollowing |
| `analyze_memory_cmdline` | Memory | Command line arguments of all processes |
| `hunt_memory_malware` | Memory | **Full autonomous 9-step triage with self-correction** |
| `analyze_disk_timeline` | Disk | log2timeline → psort super-timeline |
| `analyze_disk_fls` | Disk | Filesystem listing (allocated + deleted) |
| `analyze_registry` | Registry | SAM/SYSTEM/SOFTWARE/NTUSER hive extraction |
| `run_yara_scan` | Detection | Pattern matching with 8 built-in APT rules |
| `verify_hashes` | Integrity | SHA-256/MD5 compute or audit against known-good |
| `correlate_findings` | Analysis | Memory ↔ disk cross-reference, fileless/timestomp detection |

---

## Self-Correction Demo

The memory agent's self-correction sequence (visible in terminal output):

```
[*] Claude iteration 3/10
  -> Tool: hunt_memory_malware

[MemoryAgent] Starting autonomous memory triage...
  ~ [MemoryAgent] vol_windows_info     | 805ms | INFERRED
  ~ [MemoryAgent] vol_windows_pslist   | 31s   | INFERRED
    ↳ DELTA: pslist returned only header — rootkit DKOM confirmed
  ~ [MemoryAgent] analyze_memory_netscan | 30s  | INFERRED
  ✓ [MemoryAgent] vol_windows_malfind  | 882ms | CONFIRMED
    ↳ DELTA: Empty malfind on 90+ process system = VAD hook = rootkit IOC
  ✓ [MemoryAgent] vol_windows_cmdline  | 890ms | CONFIRMED
    ↳ DELTA: Empty cmdline = process args hidden by rootkit
  ~ [MemoryAgent] vol_windows_svcscan  | 1.2s  | INFERRED
  ✓ [MemoryAgent] psxview_diff         | 2.1s  | CONFIRMED
    ↳ DELTA: DKOM confirmed: 87 processes hidden from pslist, visible in psscan
  ~ [MemoryAgent] hollowprocesses      | 1.8s  | INFERRED
  ~ [MemoryAgent] vol_windows_dlllist  | 950ms | INFERRED
```

---

## Evidence Integrity

All operations are **read-only** by architectural enforcement:

- Volatility is called with `-f path` — read-only file access
- TSK tools are read-only by design
- No write, delete, or modify operations exist in the tool registry
- `exec.Command("vol", args...)` — not `exec.Command("bash", "-c", input)`
- SHA-256 + MD5 hashes computed at triage start, verified at end
- Spoliation test: hash before/after full triage — identical

---

## Project Structure

```
LogPoseSIFT/
├── cmd/sift-mcp/main.go                  # MCP server entry point — 12 typed tools registered
├── agents/
│   ├── orchestrator/
│   │   ├── orchestrator.go               # Claude or Gemini dual engine, 10-iteration agentic loop
│   │   └── findings_extractor.go         # Pre-triage Go parser — embeds facts before LLM starts
│   ├── memory_agent/memory.go            # 9-step autonomous memory triage with self-correction
│   ├── disk_agent/disk.go                # Evidence-type-aware disk triage, log2timeline pipeline
│   └── reasoning_logger/reasoning_logger.go  # Analyst training loop — intent/hypothesis/delta per call
├── internal/
│   ├── wrappers/                         # 7 typed tool wrappers (no raw shell)
│   │   ├── volatility.go                 # Volatility 3 memory forensics
│   │   ├── regripper.go                  # Windows registry hive extraction
│   │   ├── tsk.go                        # TSK: fls / mactime / icat
│   │   ├── bulk_extractor.go             # Carved emails, URLs, credentials
│   │   ├── foremost.go                   # File carving / recovery
│   │   ├── log2timeline.go               # Plaso super-timeline pipeline
│   │   ├── yara.go                       # YARA pattern matching (8 built-in APT rules)
│   │   ├── hashdeep.go                   # SHA-256/MD5 evidence integrity
│   │   ├── dynamic.go                    # Registry-driven tool executor
│   │   ├── executor.go                   # SafeExec — exec.Command wrapper (never bash -c)
│   │   └── helpers.go                    # Shell metachar guard, path validation, confidence consts
│   ├── registry/sift_tools.go            # 30+ tool allowlist — binary + typed fixed args only
│   ├── validator/validator.go            # Hallucination guard — CONFIRMED/INFERRED/UNVERIFIED
│   ├── correlator/correlator.go          # Disk vs memory cross-reference, DKOM/fileless detection
│   ├── logger/logger.go                  # JSONL structured audit trail per session
│   └── parsers/
│       ├── plaso_parser.go               # Plaso timeline output parser
│       └── vol_parser.go                 # Volatility output parser
├── benchmark/
│   ├── run_benchmark.sh                  # Accuracy harness — TP/FP/FN against ground truth
│   ├── ground_truth/
│   │   └── srl2018_apt_ground_truth.json # 14 documented IOCs from SRL-2018 APT case
│   └── results/                          # Scorecard JSON + Markdown per run
├── assets/                               # Screenshots for documentation
├── docs/
│   ├── architecture.md                   # Security boundaries + full data flow diagram
│   ├── accuracy_report.md                # 100% precision, 92.86% recall, 0 hallucinations
│   ├── dataset.md                        # SRL-2018 dataset documentation + reproducibility
│   ├── devpost_story.md                  # Full project story for Devpost submission
│   ├── SCREENSHOTS.md                    # 5 annotated evidence screenshots
│   ├── VIDEO.md                          # Demo video + timestamps + YouTube description
│   └── RESULT.md                         # Full live triage output from SRL-2018 run
├── data/                                 # Evidence files — not committed to git
├── logs/                                 # Session logs — JSONL + Markdown per triage run
├── logpose-ai                            # Compiled binary
├── go.mod
├── go.sum
└── README.md
```

---

## Accuracy Results

Full report: [docs/accuracy_report.md](docs/accuracy_report.md)

Tested against SRL-2018 Compromised Enterprise Network (SANS DFIR Summit dataset):

| Metric | Result |
|---|---|
| True Positives | 13 / 14 IOCs |
| False Positives | 0 |
| Hallucinations | 0 |
| Precision | **100%** |
| Recall | **92.8%** |
| Triage time | 552 seconds |

---

## License

MIT — see [LICENSE](LICENSE)

## Built For

[FIND EVIL! Hackathon](https://findevil.devpost.com/) — SANS Institute · April–June 2026
<!-- docs pass: 2026-07-03T04:33:12Z -->
