![perseus logo](assets/perseus_logo.png)

--------------------------------------------

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
* [Monitoring HTTP-Endpoint](#monitoring)

## Installation

TODO

## Usage

TODO

## Configuration

TODO

## Monitoring

TODO

## Production ready?

Yes. *perseus* runs successfully in production at [trivago](http://www.trivago.com/) and mirrors PHP packages for > 200 engineers.

Are you using *perseus* in production as well? [Open an issue and tell us](https://github.com/andygrunwald/perseus/issues/new)!

## Development

### Code structure

TODO

## Unit tests

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

### The production story

On Friday, the 6th of Jan. 2017 I had [a motivating chat](https://twitter.com/andygrunwald/status/817449096562753536) with my colleague [Matthias](https://github.com/mre).
I presented him my idea about *perseus*, mentioned multiple ideas what can be built in and asked him about a challenge I faced about the software/project architecture.
At the end of this chat he said: 

> And when you are done, you put it in production [at trivago]. After the server started, you have max. 5h to get all bug fixes done. Then it needs to run and serve us and our packages.

I just answered: 

> Deal!.

And in the end: **We did it.** And it works out. Challenges keep us motivated!

## Credits

The perseus logo was created by @mre.  
The original Gopher was designed by [Renee French](http://reneefrench.blogspot.com/).  
Go Gopher vector illustration by [Hugo Arganda](http://about.me/argandas) (@argandas)
Hosted at the [gopher-vector repository](https://github.com/golang-samples/gopher-vector).
The Medusa vector art was adjusted from [Amanda Downs](https://thenounproject.com/search/?q=medusa&i=22849) work from the Noun Project.
The perseus font is called [Dalek](http://www.dafont.com/de/dalek.font) created by [K-Type](http://www.k-type.com/).
