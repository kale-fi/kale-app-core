# KaleFi Deployment Guide

## Hetzner Deployment

This directory contains configuration files for deploying KaleFi validators on Hetzner Cloud.

### Cost Estimate

**KaleFi Hetzner deployment: €17.34/month for 5 validators**

### Hardware Requirements

Each validator uses a lightweight configuration:
- 2 GB RAM
- 20 GB SSD storage
- 1 vCPU

### Deployment Instructions

1. Create a Hetzner Cloud account and set up your API token
2. Provision 5 CX21 instances (2 GB RAM, 1 vCPU, 20 GB SSD)
3. Install Docker and Docker Compose on each instance
4. Clone this repository on each instance
5. Create the required data directories:
   ```bash
   sudo mkdir -p /var/lib/kalefi/validator-{1,2,3,4,5}
   sudo chown -R $(id -u):$(id -g) /var/lib/kalefi
   ```
6. Start the validator:
   ```bash
   cd deploy
   docker-compose -f hetzner-docker-compose.yml up -d kalefi-validator-1
   ```

### Monitoring

The validators expose the following ports:
- RPC: 26657 (and subsequent ports for other validators)
- P2P: 26656 (and subsequent ports for other validators)
- REST API: 1317 (and subsequent ports for other validators)

You can check the status of a validator using:
```bash
curl http://localhost:26657/status
```

### Troubleshooting

If you encounter issues, check the logs:
```bash
docker-compose -f hetzner-docker-compose.yml logs kalefi-validator-1
```
