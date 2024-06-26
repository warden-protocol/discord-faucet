# Discord Faucet

Another Discord Faucet for Cosmos ecosystems, specifically Warden but can work on
others.

## Config

| ENV            | Type   | Default                                       |
| -------------- | ------ | --------------------------------------------- |
| PORT           | string | 8081                                          |
| ENV_FILE       | string |                                               |
| TOKEN          | string |                                               |
| PURGE_INTERVAL | string | 10s                                           |
| MNEMONIC       | string |                                               |
| NODE           | string | https://rpc.buenavista.wardenprotocol.org:443 |
| CHAIN_ID       | string |                                               |
| CLI_NAME       | string | wardend                                       |
| ACCOUNT_NAME   | string | faucet                                        |
| DENOM          | string | uward                                         |
| AMOUNT         | string | 10000000                                      |
| FEES           | string | 25uward                                       |
| TX_RETRY       | int    | 10                                            |
| COOLDOWN       | string | 10s                                           |
