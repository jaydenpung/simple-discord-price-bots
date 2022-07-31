# Simple Discord Price Bots

This project is based on [discord-stock-ticker](https://github.com/jaydenpung/simple-discord-price-bots). Instead of starting an api server that can be called to manage bots, this project lets you define the bot in a json file, and then simply run the binary file.

The goal is to keep this project as simple as possible for any rookies to modify the code easily to achieve whatever they need.

[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](https://github.com/tterb/atomic-design-ui/blob/master/LICENSEs)
[![GitHub last commit](https://img.shields.io/github/last-commit/jaydenpung/simple-discord-price-bots.svg?style=flat)](https://github.com/jaydenpung/simple-discord-price-bots/pulse)
[![GitHub stars](https://img.shields.io/github/stars/jaydenpung/simple-discord-price-bots.svg?style=social&label=Star)](https://github.com/jaydenpung/simple-discord-price-bots/pulse)
[![GitHub watchers](https://img.shields.io/github/watchers/jaydenpung/simple-discord-price-bots.svg?style=social&label=Watch)](https://github.com/jaydenpung/simple-discord-price-bots/pulse)

## Run

1) Download [the latest release](https://github.com/jaydenpung/simple-discord-price-bots/releases/)

2) Extract the compressed release to get files including the release binary `simple-discord-price-bots`

3) Add a folder `bots` to where the release binary is.

4) Download the [sample.json](https://github.com/jaydenpung/simple-discord-price-bots/blob/master/bots/sample.json) into `bots/sample.json`

5) Rename the `sample.json` to `anyname.json`, and modify the content based on [Bot Configuration](#Bot-Configuration)

6) Run the `simply-discord-price-bots` and that's it!

### Bot Configuration

```json
{
  "name": "bitcoin",                                # string: name of the crypto from coingecko
  "crypto": true,                                   # bool: always true for crypto
  "ticker": "1) BTC",                               # string/OPTIONAL: overwrites display name of bot
  "color": true,                                    # bool/OPTIONAL: requires nickname
  "decorator": "@",                                 # string/OPTIONAL: what to show instead of arrows
  "currency": "aud",                                # string/OPTIONAL: alternative curreny
  "currency_symbol": "AUD",                         # string/OPTIONAL: alternative curreny symbol
  "pair": "binancecoin",                            # string/OPTIONAL: pair the coin with another coin, replaces activity section
  "pair_flip": true,                                # bool/OPTIONAL: show <pair>/<coin> rather than <coin>/<pair>
  "activity": "Hello;Its;Me",                       # string/OPTIONAL: list of strings to show in activity section
  "decimals": 3,                                    # int/OPTIONAL: set number of decimal places
  "nickname": true,                                 # bool/OPTIONAL: display information in nickname vs activity
  "frequency": 10,                                  # int/OPTIONAL: seconds between refresh
  "discord_bot_token": "xxxxxxxxxxxxxxxxxxxxxxxx"   # string: dicord bot token
}
```
