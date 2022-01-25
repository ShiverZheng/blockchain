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