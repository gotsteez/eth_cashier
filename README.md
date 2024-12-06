# ETH Cashier
Allows a user to send ETH, and keep that ETH in USD value. Users can then withdraw the balance in USD any time

# Setup
1. Fill out the variables in `configs/.env.example`
2. Run `anvil` to start a local eth testnet. More about that [here](https://book.getfoundry.sh/anvil/)
3. To run the server, run `go run main/main.go`

# Running
I used insomnia to test the HTTP routes. Will show example HTTP requests here

## New User
Method: `POST`
URL: `localhost:8080/newUser`
Example Response:
```
{
	"user": "1d214ab9-0878-4c61-9f51-122da3155fac",
	"walletPublicKey": "0x51075E7fE9c1FF64bb3e96db6879e0A6320f952A"
}
```

## Get User Info
Method: `POST`
URL: `localhost:8080/user`
Example Request Body
```
{
    "user": "1d214ab9-0878-4c61-9f51-122da3155fac"
}
```
Example Response
```
{
	"user": "1d214ab9-0878-4c61-9f51-122da3155fac",
	"balance": 0,
	"walletPublicKey": "0x51075E7fE9c1FF64bb3e96db6879e0A6320f952A"
}
```

## Check User
Description: Checks if they user has sent any money to the eth wallet
Method: `POST`
URL: `localhost:8080/check`
Example Request Body
```
{
    "user": "1d214ab9-0878-4c61-9f51-122da3155fac"
}
```
Example Response
```
{
	"balance": 4047.266327139078
}
```

## Withdraw
Description: Withdraws to a user wallet. To verify, check the balance of the user and admin wallet after.
Method: `POST`
URL: `localhost:8080/withdraw`
Example Request Body
```
{
    "user": "1d214ab9-0878-4c61-9f51-122da3155fac",
		"wallet": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
		"amount": 2000
}
```
Example Response
```
{
	"balance": 2047.266327139078
}
```

# NOTES
- Private key is not actually encrypted
