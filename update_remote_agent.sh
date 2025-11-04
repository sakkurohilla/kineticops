#!/bin/bash

# Update remote agent with fixed version
echo "Copying fixed agent to remote machine..."

# Copy the fixed agent binary
scp /home/akash/kineticops/agent/kineticops-agent rohilla@192.168.1.100:/tmp/

# SSH to remote machine and update
ssh rohilla@192.168.1.100 << 'EOF'
sudo systemctl stop kineticops-agent
sudo cp /tmp/kineticops-agent /opt/kineticops-agent/kineticops-agent
sudo chmod +x /opt/kineticops-agent/kineticops-agent
sudo systemctl start kineticops-agent
sudo systemctl status kineticops-agent
EOF

echo "Agent updated. Check logs in 30 seconds for CPU/memory/disk data."