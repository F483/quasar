[![Build Status](https://travis-ci.org/F483/quasar.svg)](https://travis-ci.org/F483/quasar)
[![Issues](https://img.shields.io/github/issues/f483/quasar.svg)](https://github.com/f483/quasar/issues)
[![Go Report Card](https://goreportcard.com/badge/github.com/f483/quasar)](https://goreportcard.com/report/github.com/f483/quasar)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/f483/quasar/master/LICENSE)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/f483/quasar)
[![Donate PayPal](https://img.shields.io/badge/Donate-PayPal-ff69b4.svg)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=fabian%2ebarkhau%40gmail%2ecom&lc=DE&item_name=https%3a%2f%2fgithub%2ecom%2fF483%2fquasar&no_note=0&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donateCC_LG%2egif%3aNonHostedGuest)
[![Donate Bitcoin](https://img.shields.io/badge/Donate-Bitcoin-ff69b4.svg)](https://blockchain.info/address/1PWF7UH1bFqSirC47qCtUZyuBexxZFTXDb)
[![Avaivable For Hire](https://img.shields.io/badge/Available-For_Hire-ff69b4.svg)](https://f483.github.io)


# Quasar

Go implementation of the [quasar protocol](https://www.microsoft.com/en-us/research/wp-content/uploads/2008/02/iptps08-quasar.pdf).


# Simulation Statistics

## Subscription false positive rate.

The more saturated the root subscription bloom filter becomes, the more
false positives.

TODO how to test larger filter size traffic vs larger false positive
traffic?

![Subscription False Positive rate](https://github.com/f483/quasar/raw/master/_simulation/subfprate.png)

## Topic subscription distribution.

A long tail power law is assumed for the distribution of topic
subscriptions, loosly based on twitter data.

![Subscription Distribution](https://github.com/f483/quasar/raw/master/_simulation/subdistribution.png)


## Testing Setup

The simulation uses a mock overlay network. Peers connected randomly to
one another. This allows quick isolated testing of the quasar protocol.

It does not allow testing of: Peer discovery, churn behaviour, NAT traversal or behaviour of different overlay networks.


