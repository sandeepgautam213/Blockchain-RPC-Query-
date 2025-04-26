# Blockchain-RPC-Query-


# TRON RPC Wallet Explorer (GoLang)

This project is a simple Golang service that connects to a TRON QuickNode RPC endpoint.  
It provides APIs to:

- Fetch the current balance of a wallet address.
- List all payer addresses (who sent funds).
- List all beneficiary addresses (to whom this address sent funds).

✅ Built using **only RPC methods**, **no third-party indexing services**!

---

---

## ⚙️ Prerequisites

- Golang 1.20+ installed
- QuickNode account + TRON RPC URL
- Internet connection

---

## 📥 Setup Instructions

1. **Clone this repository**

```bash
git clone https://github.com/sandeepgautam213/Blockchain-RPC-Query-.git
cd Blockchain-RPC-Query-

2. **Initialize Go modules**
go mod tidy

3. **Create .env**
touch .env

Inside .env add : 
RPC_URL=https://your-quicknode-tron-url/jsonrpc

4 Run 
go run main.go

API EndPoints 
1. Get Current Balance
URL: /currentBalance

Method: GET

Query Params:

address - TRON wallet address (starts with T)

Example Request:
curl "http://localhost:8080/currentBalance?address=TF79M4ikfNwF7b3AuTSzEEC1wWuwZzUwXy"

2. Get List of Payer Addresses
URL: /payerAddress

Method: GET

Query Params:

address - TRON wallet address

Example : 
curl "http://localhost:8080/payerAddress?address=TF79M4ikfNwF7b3AuTSzEEC1wWuwZzUwXy"

3. Get List of Beneficiary Addresses
URL: /listOfBeneficiary

Method: GET

Query Params:

address - TRON wallet address
curl "http://localhost:8080/listOfBeneficiary?address=TF79M4ikfNwF7b3AuTSzEEC1wWuwZzUwXy"



