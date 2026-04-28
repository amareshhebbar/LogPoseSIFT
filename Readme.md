
LogPoseSIFT/
├── cmd/
│   └── sift-mcp/          # Your Golang main.go for the custom MCP server
├── internal/
│   ├── wrappers/          # Safe Golang API wrappers for SIFT CLI tools (volatility, plaso)
│   └── parsers/           # Logic to turn messy terminal output into clean JSON
├── agents/
│   ├── memory_agent/      # AI logic for RAM analysis
│   ├── disk_agent/        # AI logic for hard drive timelines
│   └── orchestrator/      # The primary node that routes tasks between agents
├── data/                  # Drop your sample disk images here (ADD TO .gitignore!)
├── logs/                  # Crucial: Where your execution traces are saved for the judges
├── docs/
│   └── architecture.md    # Required architecture diagram will go here
├── .gitignore             # Ignore large /data files and binaries
├── go.mod                 # Go dependencies
└── README.md              # Project story, build instructions, and accuracy report

### 1. How to verify the connection is working
The easiest way to prove the folders are linked is to do a quick test:

1. **On your Fedora machine**, open a file explorer or terminal and create a dummy file in the project folder:
   ```bash
   touch ~/hackathon/LogPoseSIFT/hello_vm.txt
   ```
2. **Look inside your SIFT VM**, open the terminal there, and check the folder:
   ```bash
   ls ~/LogPoseSIFT
   ```
If you see `hello_vm.txt` sitting next to your `cmd`, `agents`, and `go.mod` files, your developer bridge is 100% active. (You can delete that dummy file afterward)


**The Reboot Routine (For your next coding session):**
When you sit down tomorrow to code, you just do a quick 3-step startup routine:

1. **Start the VM:** Open VirtualBox and click Start on the SIFT Workstation.
2. **Verify the IP:** Open the VM terminal and run `ip a`. (Because you are on a Bridged Adapter, your home router assigns the IP. It will *probably* stay `192.168.0.4`, but if your router restarts, it might change to `.5` or `.6`).
3. **Reconnect the Bridge:** Open your Fedora terminal and run your mount command with whatever IP it shows:
   ```bash
   sshfs sansforensics@192.168.0.4:/home/sansforensics/LogPoseSIFT ~/hackathon/LogPoseSIFT
   ```

no