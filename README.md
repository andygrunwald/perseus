![perseus logo](assets/perseus_logo.png)

--------------------------------------------

[![Build Status](https://travis-ci.org/andygrunwald/perseus.svg?branch=master)](https://travis-ci.org/andygrunwald/perseus)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygrunwald/perseus)](https://goreportcard.com/report/github.com/andygrunwald/perseus)
[![GoDoc](https://godoc.org/github.com/andygrunwald/perseus?status.svg)](https://godoc.org/github.com/andygrunwald/perseus)

Local git mirror for your PHP ([composer](https://getcomposer.org/)) project dependencies that works together with [Satis](https://github.com/composer/satis).

*perseus* is a successor out of and drop-in replacement for [Medusa](https://github.com/instaclick/medusa).

## Whats wrong with Medusa?

Nothing. Really.
[Medusa](https://github.com/instaclick/medusa) is a great software.
It works well for many people and companies.
Thanks to [Sebastien Armand](https://github.com/khepin), [Instaclick Inc.](https://github.com/instaclick) and all others who have helped and contributed to this project.
But it has its limitations, flaws and disadvantages like:

* Very poor documentation (as mentioned in the readme)
* Nearly no error handling (for API requests to [Packagist](https://packagist.org/), system commands like triggering git, etc.)
* Long mirror/update runs, due to sequential procedure and single threaded nature (long runtimes can ruin a fast development workflow)
* Stops the complete mirror/update run, if one package/url/composer.json is faulty and stops updating other packaging in the list
* Need to implement auxiliary processes to make it work in a bigger engineering team like self-service to add new or remove old packages, monitoring and reliabilities

*perseus* was born out of the motivation to eliminate these points.

## Features

* **Drop-in replacement** for [Medusa](https://github.com/instaclick/medusa)
* Fully documented
* Concurrency and usage of multiple threads for faster mirror/update runs
* Serious error handling
* Reporting of faulty packages or packages that can't be processed

## Installation

TODO

## Usage

TODO

## Configuration

TODO

## Development

### Unit tests

A running go installation is required to execute unit tests.
To execute them, run:

```
$ make test
```

Tip: If you plan to contribute via a Pull Request, the use of unit tests is encouraged.

## Project background

### The name "*perseus*"

Naming projects is hard.
I often struggle with this.
The name needs to be simple, "catchy" and easy to remember.

In this case it was easy.
[Medusa](https://en.wikipedia.org/wiki/Medusa) was part of the greek mythology.
I started looking in this direction and found *Perseus*.
Checkout [Perseus with the Head of Medusa](https://en.wikipedia.org/wiki/Perseus_with_the_Head_of_Medusa) for more details..

## Credits

The perseus logo was created by [@mre](https://github.com/mre).  
The original Gopher was designed by [Renee French](http://reneefrench.blogspot.com/).  
Go Gopher vector illustration by [Hugo Arganda](http://about.me/argandas) ([@argandas](https://github.com/argandas))  
Hosted at the [gopher-vector repository](https://github.com/golang-samples/gopher-vector).  
The Medusa vector art was adjusted from [Amanda Downs](https://thenounproject.com/search/?q=medusa&i=22849) work from the Noun Project.  
The perseus font is called [Dalek](http://www.dafont.com/de/dalek.font) created by [K-Type](http://www.k-type.com/).
