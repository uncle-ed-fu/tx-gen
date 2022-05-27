# tx-gen

very much a WIP and I will continue to make ridiculous changes as I need them without warning. This tool is not intended to be supported at all, its just a demo that I'm periodically adding tools to as I test out and debug mamaki. Not even sure if I left the commands in a work state.

tool to help spam transactions and query the chain. also, pay to make pictures on celestia available

### install
```sh
go install
```

### usuage
generate a bunch of keys. See the accounts package. It should spit out a "controller" account. Send funds to that account, and then you can use the next command to disperse those funds further to different accounts
```
tx-gen init
```

send funds to all those accounts you just generated.
```
tx-gen fund --node tcp://my.node.ip.here:26657
```

query basic onchain data and save it to a json

```
tx-gen history summary --start 3000 --node tcp://my.node.ip.here:26657
```

still need to debug why the signature fails sometimes, but the data ends up on chain and you still have to pay for gas
```
tx-gen photo --node tcp://my.node.ip.here:26657 --path /path/to/photo/photo.jpg
```
generate and send a bunch of pay for data txs with random data. This command should actually work without tinkering unlike the above commands.
```
tx-gen pfd  --node tcp://159.203.18.242:26657 --size 400000 --actors 20
```
