# Circle Layer Blockchain Node

This project provides installation, running, and maintenance capabilities for **Circle Layer validator nodes** supporting the Circle Layer blockchain ecosystem. The consensus structure uses Delegated Proof of Stake (DPoS) with Circle Layer's implementation of go-ethereum and system contracts. This repository is actively maintained with regular updates for enhanced functionality and stability.

## 🌐 Circle Layer Network

**Official Website**: [https://circlelayer.com](https://circlelayer.com)

**Network Information**:
- **Chain ID**: 28525
- **Currency**: CLAYER
- **Block Time**: 3 seconds
- **Finality**: 1-3 seconds

**Testnet Resources**:
- **RPC Endpoint**: [https://testnet-rpc.circlelayer.com](https://testnet-rpc.circlelayer.com)
- **WebSocket**: wss://testnet-rpc.circlelayer.com
- **Block Explorer**: [https://explorer-testnet.circlelayer.com](https://explorer-testnet.circlelayer.com)
- **Faucet**: [https://faucet.circlelayer.com](https://faucet.circlelayer.com)
- **Faucet API**: [https://faucet-api.circlelayer.com](https://faucet-api.circlelayer.com)

## System Requirements

**Operating System:** Ubuntu >= 20.04 LTS

**RAM:** 8GB minimum, 32GB recommended

**Persistent Storage:** 25GB minimum, 100GB high-speed SSD recommended

**Note regarding use of GPUs -** GPUs are primarily used in POW consensus chains. Being a DPos CLAYER chain has not only more TPS and fast block production but also doesn't need a GPU altogether for its purpose.

## How to become a validator
To back the CLAYER blockchain you can become a validator. Full flow to become a validator, you must:

* Install this package **([See Installation](#installation))**

* Download your newly created validator wallet from your server and import it into your metamask or preferred wallet. Fund this account with the appropriate CLAYER Coins needed to become a validator. Example command to download the wallet on your local PC. Only works for UNIX-based OSes or on any environment that can run the OpenSSH package:
```bash
  scp -r root@<server_ip>:/root/testnet-core-blockchain/chaindata/node1/keystore ./
  scp root@<server_ip>:/root/testnet-core-blockchain/chaindata/node1/pass.txt ./
```

* On your server, start the node that you just installed **([See Usage/Example](#usageexamples))**

* Once the node is started and confirmation is seen on your terminal, open the interactive console by attaching tmux session **([See Usage/Example](#usageexamples))**

* Once inside the interactive console, you'll see "IMPORTED TRANSACTION OBJECTS" and "age=<some period like 6d5hr or 5mon 3weeks>". You need to wait until the "unauthorized validator" warning starts to pop up on the console. 

* Once "unauthorized validators" warning shows up, go to https://staking.circlelayer.com/ and click "Become a validator". Fill in all the details in the form, in the "Fee address" field enter the validator wallet address that you imported into your metamask. Proceed further

* Once the last step is done, you'll see a "🔨 mined potential block" message on the interactive console. You'll also see your validator wallet as a validator on the staking page and on explorer. You should also detach from the console after the whole process is done **([See Usage/Example](#usageexamples))**

## Installation

**You must ensure that:** 
* system requirements are met with careful supervision
* the concerned server/local setup must be running 24/7 
* there is sufficient power and cooling arrangement for your machine if running a local setup 
If failed you may end up losing your stake in the blockchain and your staked coins, if any. You'll be jailed at once with no return point by the blockchain if found down/dead. You'll be responsible for chain data corruption on your node, frying up your motherboard, or damaging yourself and your surroundings. 


To install the CLAYER validator node in ubuntu linux
```bash
  sudo -i
  apt update && apt upgrade
  apt install git tar curl wget
  reboot
```
Skip the above commands if you have already updated the system and installed the necessary tools.

Connect again to your server after reboot
```bash
  sudo -i
  git clone https://github.com/CircleLayer/testnet-core-blockchain.git
  cd testnet-core-blockchain
  ./node-setup.sh --validator 1
```
After you run node-setup, follow the on-screen instructions carefully and you'll get confirmation that the node was successfully installed on your system.

**Note regarding your validator account -** While in the setup process, you'll be asked to create a new account that must be used for block mining and receiving gas rewards. You must import this account to your metamask or any preferred wallet. 
 
    
## Usage/Examples

Display help
```bash
./node-setup.sh --help
```
To create/install a validator node. Fresh first-time install
```bash
./node-setup.sh --validator 1
source ~/.bashrc
```
To run the validator node
```bash
./node-start.sh --validator
```
To create/install a RPC node. Fresh first-time install
```bash
./node-setup.sh --rpc
source ~/.bashrc
```
To run the RPC node
```bash
./node-start.sh --rpc
```
To stop the RPC node
```bash
./node-stop.sh --rpc
```
To stop the Validator node
```bash
./node-stop.sh --validator
```
To get into a running node's interactive console/tmux session 
```bash
tmux attach -t node1
```
To stop a running node or the running blockchain node 
```bash
tmux attach -t node1
```

To exit/detach from an interactive console
```text
Press CTRL & b , release both keys, and press d
```

## 📚 Documentation & Resources

**Complete Documentation**: [https://docs.circlelayer.com](https://docs.circlelayer.com)

**Developer Resources**:
- [Getting Started Guide](https://docs.circlelayer.com/getting-started/set-up-wallet)
- [Smart Contract Development](https://docs.circlelayer.com/development/writing-smart-contracts)
- [API & RPC Documentation](https://docs.circlelayer.com/apis-sdks/rpc-endpoints)
- [Validator Guide](https://docs.circlelayer.com/nodes-validation/becoming-validator)
- [Node Setup Tutorial](https://docs.circlelayer.com/nodes-validation/running-full-node)

**Network Configuration**:
- [Connect to Testnet](https://docs.circlelayer.com/getting-started/connect-testnet)
- [Use Faucet](https://docs.circlelayer.com/getting-started/use-faucet)
- [Web3 Integration](https://docs.circlelayer.com/development/web3-integration)

## 🤝 Community & Support

**Social Media**:
- **X (Twitter)**: [https://x.com/circlelayer](https://x.com/circlelayer)
- **Telegram**: [https://t.me/circlelayer](https://t.me/circlelayer)

**Development**:
- **GitHub**: [https://github.com/circlelayer](https://github.com/circlelayer)
- **Issues & Support**: [GitHub Issues](https://github.com/CircleLayer/testnet-core-blockchain/issues)

**Contact & Support**:
- **Technical Support**: [support@circlelayer.com](mailto:support@circlelayer.com)
- **General Inquiries**: [admin@circlelayer.com](mailto:admin@circlelayer.com)
- **Marketing & Partnerships**: [marketing@circlelayer.com](mailto:marketing@circlelayer.com)

**Community Guidelines**: [https://docs.circlelayer.com/community/social-media](https://docs.circlelayer.com/community/social-media)

## 🔗 Additional Resources

**For Developers**:
- [API Reference](https://docs.circlelayer.com/apis-sdks/rpc-endpoints)
- [SDK Documentation](https://docs.circlelayer.com/apis-sdks/web3-libraries)
- [Development Examples](https://docs.circlelayer.com/development/writing-smart-contracts)

**For Validators**:
- [Validator Requirements](https://docs.circlelayer.com/nodes-validation/becoming-validator)
- [Node Monitoring](https://docs.circlelayer.com/nodes-validation/node-monitoring)
- [Security Guidelines](https://docs.circlelayer.com/security/risk-warnings)

**Network Information**:
- [Architecture Overview](https://docs.circlelayer.com/architecture/genesis)
- [Consensus Mechanism](https://docs.circlelayer.com/architecture/pos-consensus)
- [Tokenomics](https://docs.circlelayer.com/governance/tokenomics)

---

**Circle Layer** - Building the future of scalable, secure blockchain infrastructure.

For more information, visit [https://circlelayer.com](https://circlelayer.com) or read our [complete documentation](https://docs.circlelayer.com).