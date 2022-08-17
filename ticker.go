package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jaydenpung/simple-discord-price-bots/utils"
	log "github.com/sirupsen/logrus"
)

type Ticker struct {
	Ticker         string   `json:"ticker"`
	Name           string   `json:"name"`
	NamePrefix     string   `json:"name_prefix"`
	Nickname       bool     `json:"nickname"`
	Frequency      int      `json:"frequency"`
	Color          bool     `json:"color"`
	Decorator      string   `json:"decorator"`
	Currency       string   `json:"currency"`
	CurrencySymbol string   `json:"currency_symbol"`
	Decimals       int      `json:"decimals"`
	Activity       string   `json:"activity"`
	Pair           string   `json:"pair"`
	PairFlip       bool     `json:"pair_flip"`
	Multiplier     int      `json:"multiplier"`
	ClientID       string   `json:"client_id"`
	Crypto         bool     `json:"crypto"`
	Token          string   `json:"discord_bot_token"`
	TwelveDataKey  string   `json:"twelve_data_key"`
	Exrate         float64  `json:"exrate"`
	Close          chan int `json:"-"`
	Api            string   `json:"api"`
	IsPercent      bool     `json:"is_percent"`
}

// label returns a human readble id for this bot
func (s *Ticker) label() string {
	var label string
	if s.Crypto {
		label = strings.ToLower(fmt.Sprintf("%s-%s", s.Name, s.Currency))
	} else {
		label = strings.ToLower(fmt.Sprintf("%s-%s", s.Ticker, s.Currency))
	}
	if len(label) > 32 {
		label = label[:32]
	}
	return label
}

func (s *Ticker) runGecko() {
	// create a new discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + s.Token)
	if err != nil {
		log.Errorf("Creating Discord session: %s", err)
		return
	}

	// get shards
	st, err := dg.GatewayBot()
	if err != nil {
		log.Errorf("Creating Discord gateway: %s", err)
		return
	}

	// shard into sessions.
	shards := make([]*discordgo.Session, st.Shards)
	for i := 0; i < st.Shards; i++ {
		shards[i], err = discordgo.New("Bot " + s.Token)
		if err != nil {
			log.Errorf("Creating Discord sharded session: %s", err)
			return
		}
		shards[i].ShardID = i
		shards[i].ShardCount = st.Shards
	}

	// open ws connections
	var errOpen error
	{
		wg := sync.WaitGroup{}
		for _, sess := range shards {
			wg.Add(1)
			go func(sess *discordgo.Session) {
				if err := sess.Open(); err != nil {
					errOpen = err
				}
				wg.Done()
			}(sess)
		}
		wg.Wait()
	}
	if errOpen != nil {
		wg := sync.WaitGroup{}
		for _, sess := range shards {
			wg.Add(1)
			go func(sess *discordgo.Session) {
				_ = sess.Close()
				wg.Done()
			}(sess)
		}
		wg.Wait()
	}

	// Get guilds for bot
	guilds, err := dg.UserGuilds(100, "", "")
	if err != nil {
		log.Errorf("Getting guilds: %s", err)
		s.Nickname = false
	}

	// If other currency, get rate
	if s.Currency != "USD" {
		log.Infof("Using %s", s.Currency)
		exData, err := utils.GetStockPrice(s.Currency + "=X")
		if err != nil {
			log.Errorf("Unable to fetch exchange rate for %s, default to USD.", s.Currency)
		} else {
			if len(exData.QuoteSummary.Results) > 0 {
				s.Exrate = exData.QuoteSummary.Results[0].Price.RegularMarketPrice.Raw * float64(s.Multiplier)
			} else {
				log.Errorf("Bad exchange rate for %s, default to USD.", s.Currency)
				s.Exrate = float64(s.Multiplier)
			}
		}
	} else {
		s.Exrate = float64(s.Multiplier)
	}

	// Set arrows if no custom decorator
	var arrows bool
	if s.Decorator == "" {
		arrows = true
	}

	setName(dg, s.label())

	log.Infof("Watching crypto price for %s", s.Name)
	ticker := time.NewTicker(time.Duration(s.Frequency) * time.Second)

	// continuously watch
	for {
		select {
		case <-s.Close:
			log.Infof("Shutting down price watching for %s", s.Name)
			return
		case <-ticker.C:
			log.Debugf("Fetching crypto price for %s", s.Name)

			var priceData utils.GeckoPriceResults
			var fmtPrice string
			var fmtChange string
			var changeHeader string
			var fmtDiffPercent string

			// get the coin price data
			priceData, err = utils.GetCryptoPrice(s.Name)

			if err != nil {
				log.Errorf("Unable to fetch crypto price for %s: %s", s.Name, err)
				continue
			}

			// Check if conversion is needed
			if s.Exrate > 1.0 {
				priceData.MarketData.CurrentPrice.USD = s.Exrate * priceData.MarketData.CurrentPrice.USD
				priceData.MarketData.PriceChangeCurrency.USD = s.Exrate * priceData.MarketData.PriceChangeCurrency.USD
			}

			// format the price changes
			fmtDiffPercent = fmt.Sprintf("%.2f", priceData.MarketData.PriceChangePercent)
			fmtChange = fmt.Sprintf("%.2f", priceData.MarketData.PriceChangeCurrency.USD)

			// Check for custom decimal places
			switch s.Decimals {
			case 0:
				fmtPrice = fmt.Sprintf("%s%.0f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 1:
				fmtPrice = fmt.Sprintf("%s%.1f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 2:
				fmtPrice = fmt.Sprintf("%s%.2f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 3:
				fmtPrice = fmt.Sprintf("%s%.3f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 4:
				fmtPrice = fmt.Sprintf("%s%.4f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 5:
				fmtPrice = fmt.Sprintf("%s%.5f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 6:
				fmtPrice = fmt.Sprintf("%s%.6f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 7:
				fmtPrice = fmt.Sprintf("%s%.7f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 8:
				fmtPrice = fmt.Sprintf("%s%.8f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 9:
				fmtPrice = fmt.Sprintf("%s%.9f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 10:
				fmtPrice = fmt.Sprintf("%s%.10f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 11:
				fmtPrice = fmt.Sprintf("%s%.11f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 12:
				fmtPrice = fmt.Sprintf("%s%.12f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			case 13:
				fmtPrice = fmt.Sprintf("%s%.13f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
			default:

				// Check for cryptos below 1c
				if priceData.MarketData.CurrentPrice.USD < 0.01 {
					priceData.MarketData.CurrentPrice.USD = priceData.MarketData.CurrentPrice.USD * 100
					if priceData.MarketData.CurrentPrice.USD < 0.00001 {
						fmtPrice = fmt.Sprintf("%.8f¢", priceData.MarketData.CurrentPrice.USD)
					} else {
						fmtPrice = fmt.Sprintf("%.6f¢", priceData.MarketData.CurrentPrice.USD)
					}
				} else if priceData.MarketData.CurrentPrice.USD < 1.0 {
					fmtPrice = fmt.Sprintf("%s%.3f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
				} else {
					fmtPrice = fmt.Sprintf("%s%.2f", s.CurrencySymbol, priceData.MarketData.CurrentPrice.USD)
				}
			}

			// calculate if price has moved up or down
			var increase bool
			if len(fmtChange) == 0 {
				increase = true
			} else if string(fmtChange[0]) == "-" {
				increase = false
			} else {
				increase = true
			}

			// set arrows based on movement
			if arrows {
				s.Decorator = "⬊"
				if increase {
					s.Decorator = "⬈"
				}
			}

			// update nickname instead of activity
			if s.Nickname {
				var displayName string
				var nickname string
				var activity string

				// override coin symbol
				if s.Ticker != "" {
					displayName = s.Ticker
				} else {
					displayName = strings.ToUpper(priceData.Symbol)
				}

				// format nickname
				if displayName == s.Decorator {
					nickname = fmtPrice
				} else {
					nickname = fmt.Sprintf("%s%s%s %s", s.NamePrefix, displayName, s.Decorator, fmtPrice)
				}

				// format activity
				if s.Pair != "" {

					// get price of target pair
					var pairPriceData utils.GeckoPriceResults
					pairPriceData, err = utils.GetCryptoPrice(s.Pair)
					if err != nil {
						log.Errorf("Unable to fetch pair price for %s: %s", s.Pair, err)
						activity = fmt.Sprintf("%s%s (%s%%)", changeHeader, fmtChange, fmtDiffPercent)
					} else {

						// set pair
						var pairPrice float64
						var pairSymbol string
						if s.PairFlip {
							pairPrice = pairPriceData.MarketData.CurrentPrice.USD / priceData.MarketData.CurrentPrice.USD
							pairSymbol = fmt.Sprintf("%s/%s", strings.ToUpper(pairPriceData.Symbol), displayName)
						} else {
							pairPrice = priceData.MarketData.CurrentPrice.USD / pairPriceData.MarketData.CurrentPrice.USD
							pairSymbol = fmt.Sprintf("%s/%s", displayName, strings.ToUpper(pairPriceData.Symbol))
						}

						// format decimals
						if pairPrice < 0.1 {
							activity = fmt.Sprintf("%.4f %s", pairPrice, pairSymbol)
						} else {
							activity = fmt.Sprintf("%.2f %s", pairPrice, pairSymbol)
						}
					}
				} else {
					if math.Abs(priceData.MarketData.PriceChangeCurrency.USD) < 0.01 {
						activity = fmt.Sprintf("%s%%", fmtDiffPercent)
					} else {
						activity = fmt.Sprintf("%s%s (%s%%)", changeHeader, fmtChange, fmtDiffPercent)
					}
				}

				if s.Activity == "marketcap" {
					marketCap := priceData.MarketData.MarketCap.USD
					activity = fmt.Sprintf("%s TVL", utils.FormatAmount(marketCap))
				}

				// Update nickname in guilds
				for _, g := range guilds {
					err = dg.GuildMemberNickname(g.ID, "@me", nickname)
					if err != nil {
						log.Errorf("Updating nickname: %s", err)
						continue
					}
					log.Debugf("Set nickname in %s: %s", g.Name, nickname)

					if s.Color {
						// change bot color
						err = setRole(dg, s.ClientID, g.ID, increase)
						if err != nil {
							log.Errorf("Color roles: %s", err)
						}
					}

					//time.Sleep(time.Duration(s.Frequency) * time.Second)
				}

				// set activity
				wg := sync.WaitGroup{}
				const ActivityTypeWatching = 3
				for _, sess := range shards {
					err = sess.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{
							{
								Name: activity,
								Type: ActivityTypeWatching,
							},
						},
						Status: "online",
					})
					if err != nil {
						log.Errorf("Unable to set activity: %s", err)
					} else {
						log.Debugf("Set activity: %s", activity)
					}
				}
				wg.Wait()

			} else {

				// format activity
				activity := fmt.Sprintf("%s %s %s%%", fmtPrice, s.Decorator, fmtDiffPercent)

				wg := sync.WaitGroup{}
				for _, sess := range shards {
					err = sess.UpdateGameStatus(0, activity)
					if err != nil {
						log.Errorf("Unable to set activity: %s", err)
					} else {
						log.Debugf("Set activity: %s", activity)
					}
				}
				wg.Wait()
			}
		}
	}
}

func (s *Ticker) runPony() {
	// create a new discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + s.Token)
	if err != nil {
		log.Errorf("Creating Discord session: %s", err)
		return
	}

	// get shards
	st, err := dg.GatewayBot()
	if err != nil {
		log.Errorf("Creating Discord gateway: %s", err)
		return
	}

	// shard into sessions.
	shards := make([]*discordgo.Session, st.Shards)
	for i := 0; i < st.Shards; i++ {
		shards[i], err = discordgo.New("Bot " + s.Token)
		if err != nil {
			log.Errorf("Creating Discord sharded session: %s", err)
			return
		}
		shards[i].ShardID = i
		shards[i].ShardCount = st.Shards
	}

	// open ws connections
	var errOpen error
	{
		wg := sync.WaitGroup{}
		for _, sess := range shards {
			wg.Add(1)
			go func(sess *discordgo.Session) {
				if err := sess.Open(); err != nil {
					errOpen = err
				}
				wg.Done()
			}(sess)
		}
		wg.Wait()
	}
	if errOpen != nil {
		wg := sync.WaitGroup{}
		for _, sess := range shards {
			wg.Add(1)
			go func(sess *discordgo.Session) {
				_ = sess.Close()
				wg.Done()
			}(sess)
		}
		wg.Wait()
	}

	// Get guilds for bot
	guilds, err := dg.UserGuilds(100, "", "")
	if err != nil {
		log.Errorf("Getting guilds: %s", err)
		s.Nickname = false
	}

	// If other currency, get rate
	if s.Currency != "USD" {
		log.Infof("Using %s", s.Currency)
		exData, err := utils.GetStockPrice(s.Currency + "=X")
		if err != nil {
			log.Errorf("Unable to fetch exchange rate for %s, default to USD.", s.Currency)
		} else {
			if len(exData.QuoteSummary.Results) > 0 {
				s.Exrate = exData.QuoteSummary.Results[0].Price.RegularMarketPrice.Raw * float64(s.Multiplier)
			} else {
				log.Errorf("Bad exchange rate for %s, default to USD.", s.Currency)
				s.Exrate = float64(s.Multiplier)
			}
		}
	} else {
		s.Exrate = float64(s.Multiplier)
	}

	// Set arrows if no custom decorator
	var arrows bool
	if s.Decorator == "" {
		arrows = true
	}

	setName(dg, s.label())

	log.Infof("Watching crypto price for %s", s.Name)
	ticker := time.NewTicker(time.Duration(s.Frequency) * time.Second)

	// continuously watch
	for {
		select {
		case <-s.Close:
			log.Infof("Shutting down price watching for %s", s.Name)
			return
		case <-ticker.C:
			log.Debugf("Fetching crypto price for %s", s.Name)

			var priceData utils.PonyPriceResults
			var fmtPrice string
			var fmtChange string
			var fmtDiffPercent string

			// get the coin price data
			priceData, err = utils.GetPonyData(s.Name)

			if err != nil {
				log.Errorf("Unable to fetch crypto price for %s: %s", s.Name, err)
				continue
			}

			var price float64
			if s.Ticker == "NAV" {
				price = priceData.NetAssetValue
			} else if s.Ticker == "APY" {
				price = priceData.AnnualPercentageYield
			} else if s.Ticker == "TVL" {
				price = priceData.TotalValueLocked
			}

			// Check if conversion is needed
			if s.Exrate > 1.0 {
				price = s.Exrate * price
			}

			// Check for custom decimal places
			switch s.Decimals {
			case 0:
				fmtPrice = fmt.Sprintf("%s%.0f", s.CurrencySymbol, price)
			case 1:
				fmtPrice = fmt.Sprintf("%s%.1f", s.CurrencySymbol, price)
			case 2:
				fmtPrice = fmt.Sprintf("%s%.2f", s.CurrencySymbol, price)
			case 3:
				fmtPrice = fmt.Sprintf("%s%.3f", s.CurrencySymbol, price)
			case 4:
				fmtPrice = fmt.Sprintf("%s%.4f", s.CurrencySymbol, price)
			case 5:
				fmtPrice = fmt.Sprintf("%s%.5f", s.CurrencySymbol, price)
			case 6:
				fmtPrice = fmt.Sprintf("%s%.6f", s.CurrencySymbol, price)
			case 7:
				fmtPrice = fmt.Sprintf("%s%.7f", s.CurrencySymbol, price)
			case 8:
				fmtPrice = fmt.Sprintf("%s%.8f", s.CurrencySymbol, price)
			case 9:
				fmtPrice = fmt.Sprintf("%s%.9f", s.CurrencySymbol, price)
			case 10:
				fmtPrice = fmt.Sprintf("%s%.10f", s.CurrencySymbol, price)
			case 11:
				fmtPrice = fmt.Sprintf("%s%.11f", s.CurrencySymbol, price)
			case 12:
				fmtPrice = fmt.Sprintf("%s%.12f", s.CurrencySymbol, price)
			case 13:
				fmtPrice = fmt.Sprintf("%s%.13f", s.CurrencySymbol, price)
			default:

				// Check for cryptos below 1c
				if price < 0.01 {
					price = price * 100
					if price < 0.00001 {
						fmtPrice = fmt.Sprintf("%.8f¢", price)
					} else {
						fmtPrice = fmt.Sprintf("%.6f¢", price)
					}
				} else if price < 1.0 {
					fmtPrice = fmt.Sprintf("%s%.3f", s.CurrencySymbol, price)
				} else {
					fmtPrice = fmt.Sprintf("%s%.2f", s.CurrencySymbol, price)
				}
			}

			if s.IsPercent {
				fmtPrice = fmt.Sprint(fmtPrice, "%")
			}

			// calculate if price has moved up or down
			var increase bool
			if len(fmtChange) == 0 {
				increase = true
			} else if string(fmtChange[0]) == "-" {
				increase = false
			} else {
				increase = true
			}

			// set arrows based on movement
			if arrows {
				s.Decorator = "⬊"
				if increase {
					s.Decorator = "⬈"
				}
			}

			// update nickname instead of activity
			if s.Nickname {
				var displayName string
				var nickname string
				var activity string

				// override coin symbol
				if s.Ticker != "" {
					displayName = s.Ticker
				} else {
					displayName = strings.ToUpper("PONY")
				}

				// format nickname
				if displayName == s.Decorator {
					nickname = fmtPrice
				} else {
					nickname = fmt.Sprintf("%s%s %s", displayName, s.Decorator, fmtPrice)
				}

				// Update nickname in guilds
				for _, g := range guilds {
					err = dg.GuildMemberNickname(g.ID, "@me", nickname)
					if err != nil {
						log.Errorf("Updating nickname: %s", err)
						continue
					}
					log.Debugf("Set nickname in %s: %s", g.Name, nickname)

					if s.Color {
						// change bot color
						err = setRole(dg, s.ClientID, g.ID, increase)
						if err != nil {
							log.Errorf("Color roles: %s", err)
						}
					}

					//time.Sleep(time.Duration(s.Frequency) * time.Second)
				}

				// set activity
				wg := sync.WaitGroup{}
				const ActivityTypeWatching = 3
				for _, sess := range shards {
					err = sess.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{
							{
								Name: activity,
								Type: ActivityTypeWatching,
							},
						},
						Status: "online",
					})
					if err != nil {
						log.Errorf("Unable to set activity: %s", err)
					} else {
						log.Debugf("Set activity: %s", activity)
					}
				}
				wg.Wait()

			} else {

				// format activity
				activity := fmt.Sprintf("%s %s %s%%", fmtPrice, s.Decorator, fmtDiffPercent)

				wg := sync.WaitGroup{}
				for _, sess := range shards {
					err = sess.UpdateGameStatus(0, activity)
					if err != nil {
						log.Errorf("Unable to set activity: %s", err)
					} else {
						log.Debugf("Set activity: %s", activity)
					}
				}
				wg.Wait()
			}
		}
	}
}

// setName will update a bots name
func setName(session *discordgo.Session, name string) {

	user, err := session.User("@me")
	if err != nil {
		log.Errorf("Getting bot user: %s", err)
		return
	}

	if user.Username == name {
		log.Debugf("Username already matches: %s", name)
		return
	}

	_, err = session.UserUpdate(name, "")
	if err != nil {
		log.Errorf("Updating bot username: %s", err)
		return
	}

	log.Debugf("%s changed to %s", user.Username, name)
}

// setRole changes color roles based on change
func setRole(session *discordgo.Session, id, guild string, increase bool) error {
	var redRole string
	var greeenRole string

	// get the roles for color changing
	roles, err := session.GuildRoles(guild)
	if err != nil {
		return err
	}

	// find role ids
	for _, r := range roles {
		if r.Name == "tickers-red" {
			redRole = r.ID
		} else if r.Name == "tickers-green" {
			greeenRole = r.ID
		}
	}

	// make sure roles exist
	if len(redRole) == 0 || len(greeenRole) == 0 {
		return errors.New("unable to find roles for color changes")
	}

	// assign role based on change
	if increase {
		err = session.GuildMemberRoleRemove(guild, id, redRole)
		if err != nil {
			return err
		}
		err = session.GuildMemberRoleAdd(guild, id, greeenRole)
		if err != nil {
			return err
		}
	} else {
		err = session.GuildMemberRoleRemove(guild, id, greeenRole)
		if err != nil {
			return err
		}
		err = session.GuildMemberRoleAdd(guild, id, redRole)
		if err != nil {
			return err
		}
	}

	return nil
}
