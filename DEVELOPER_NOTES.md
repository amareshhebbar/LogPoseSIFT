
# LogPoseSIFT: Developer Playbook

This document outlines the core directory structure, environment verification steps, and the daily startup routine for working on the LogPoseSIFT project via the Fedora-to-VM SSHFS bridge.

## 1. Project Architecture Map

This is the map of where code lives and what it does. 

```text
LogPoseSIFT/
├── cmd/
│   └── sift-mcp/          # Entry point: The Golang main.go for the custom MCP server
├── internal/
│   ├── wrappers/          # Security Boundary: Safe Golang API wrappers for SIFT CLI tools (e.g., Volatility, Plaso)
│   └── parsers/           # Formatting: Logic to convert messy terminal output into structured JSON
├── agents/
│   ├── memory_agent/      # AI Logic: Prompting and routing for RAM analysis
│   ├── disk_agent/        # AI Logic: Prompting and routing for hard drive timelines
│   └── orchestrator/      # Main Node: Routes tasks between specialized agents and enforces iteration caps
├── data/                  # Ignored by Git: Drop sample disk images and raw memory dumps here
├── logs/                  # Required Deliverable: Execution traces, token usage, and agent communication logs
├── docs/                  # Required Deliverables: Architecture diagrams, accuracy reports, etc.
├── .gitignore             # Prevents massive forensic data files from breaking Git
├── go.mod                 # Golang dependency tracking
└── README.md              # Project story, build instructions, and hackathon checklist
```

---

## 2. The Daily Startup Routine

When starting a new coding session, the SSHFS bridge must be re-established. Follow this 3-step routine.

### Step 1: Boot the SIFT Environment
1. Open VirtualBox on the Fedora host.
2. Select the **SIFT Workstation** VM and click **Start**.
3. Log in using the standard credentials (`sansforensics` / `forensics`).

### Step 2: Grab the Current IP Address
Because the VM is running on a Bridged Adapter, your local router assigns its IP. It may change if the router restarts.
1. Open a terminal inside the SIFT VM.
2. Run the following command:
   ```bash
   ip a
   ```
3. Locate the `inet` address under the primary network interface (e.g., `enp0s17` or `eth0`). It will look similar to `192.168.0.4`.

### Step 3: Mount the Developer Bridge
1. Open a terminal on your Fedora host machine.
2. Mount the remote VM directory to your local workspace using the IP address found in Step 2:
   ```bash
   sshfs sansforensics@<CURRENT_VM_IP>:/home/sansforensics/LogPoseSIFT ~/hackathon/LogPoseSIFT
   ```
3. Enter the password (`forensics`) when prompted. You can now open `~/hackathon/LogPoseSIFT` in your local IDE.

---

## 3. Verifying the Connection

If you suspect the bridge has dropped or did not mount correctly, perform this quick two-step check:

**Action on Fedora Host:**
Create a test file in your local directory.
```bash
touch ~/hackathon/LogPoseSIFT/connection_test.txt
```

**Action in SIFT VM:**
Check if the file immediately appears in the remote directory.
```bash
ls ~/LogPoseSIFT
```

If `connection_test.txt` is visible in the VM terminal, the bridge is fully active. You can then safely delete the test file.