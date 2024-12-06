# Send money to user account
Format
```
cast send --from 0xSenderAddress --private-key 0xYourPrivateKey 0xRecipientAddress --value 1000000000000000000
```

Example:
```
cast send --from 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 0x51075E7fE9c1FF64bb3e96db6879e0A6320f952A --value 1000000000000000000
```

cast balance <address> --rpc-url http://localhost:8545
My Wallet
cast balance 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 --rpc-url http://localhost:8545
User wallet on site
cast balance 0x54b5d0C56c996384e7De4c1087Cbe6912A13d5d5 --rpc-url http://localhost:8545
User private wallet
cast balance 0x70997970C51812dc3A010C7d01b50e0d17dc79C8 --rpc-url http://localhost:8545
