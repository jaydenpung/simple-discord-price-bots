# Simple Discord Price Bots

This project is based on [discord-stock-ticker](https://github.com/rssnyder/discord-stock-ticker). Instead of starting an api server that can be called to manage bots, this project lets you define the bot in a json file, and then simply run the binary file.

The goal is to keep this project as simple as possible for any rookies to modify the code easily to achieve whatever they need.

## Run

1) Rename `bots/sample.json` to `anyname.json`

2) Run `go build`, this will create a `simple-dicord-price-bots` executable file based on your current os

3) Run the `simply-discord-price-bots` and that's it!
