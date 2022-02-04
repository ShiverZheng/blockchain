## Blockchain

Here is the [tutorial](https://mycoralhealth.medium.com/code-your-own-blockchain-in-less-than-200-lines-of-go-e296282bcffc)


### [Basic blockchain](https://github.com/ShiverZheng/blockchain/tree/basic)

Get Blockchain
```shell
$ curl localhost:8080
```

Set Blockchain
```shell
$ curl -d '{ "BPM": 55  }' -X POST localhost:8080
```

### [Networking](https://github.com/ShiverZheng/blockchain/tree/networking)

Serve TCP server

```shell
$ go run .
```

Open a new terminal, typing `nc localhost 9000` and then follow the prompts.

### [Proof-of-Work](https://github.com/ShiverZheng/blockchain/tree/proof-of-work)

Set the difficulty in `blockchain.go`

```shell
$ go run .
```

```shell
$ curl -d '{ "BPM": 66 }' -X POST localhost:9000
```

Wait for the hash value that meets the requirements to be calculated.

### Asymmetric encryption:  

1. The receiver generates a pair of keys, **private key**, **public key**
2. The sender encrypts the data with the received public key and sends it to the receiver
3. The receiver sends the public key to the sender
4. The receiver receives the data, it decrypts it with its own private key

> In the asymmetric algorithm, the data encrypted by the public key can only be decrypted with the corresponding private key, and the private key is only known by the receiver, thus ensuring the security of data transmission.

### [Proof-of-Stake](https://github.com/ShiverZheng/blockchain/tree/proof-of-stake)

Serve TCP server

```shell
$ go run .
```
> We get our genesisBlock printed on our console

Open a new terminal, typing `nc localhost 9000` and then follow the prompts.

We’re then prompted to add a token balance to stake. Enter the number of tokens you want that validator to stake. Then input a pulse rate for that validator.

Because we can have many validators, let’s do the same thing with another terminal window.

**Watch your first terminal as you’re adding new terminals. We see validators get assigned addresses and we get a list of validators each time a new one is added!**

Wait a little while...

Check out your new terminals. The program is spending some time picking a winner. 

And then ... 

A winner is chosen!
> We can verify the validator’s address by comparing it to the list of validators printed in our main terminal

Wait a little while again...

We see our new blockchain broadcast to all our terminals, with our winning validator’s block containing his BPM in the newest block!