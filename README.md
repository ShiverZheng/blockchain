## Blockchain

Here is the [tutorial](https://mycoralhealth.medium.com/code-your-own-blockchain-in-less-than-200-lines-of-go-e296282bcffc)


### [Basic blockchain](https://github.com/ShiverZheng/blockchain)

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