![perseus logo](assets/perseus_logo.png)

--------------------------------------------

[![Build Status](https://travis-ci.org/andygrunwald/perseus.svg?branch=master)](https://travis-ci.org/andygrunwald/perseus)
[![Go Report Card](https://goreportcard.com/badge/github.com/andygrunwald/perseus)](https://goreportcard.com/report/github.com/andygrunwald/perseus)
[![GoDoc](https://godoc.org/github.com/andygrunwald/perseus?status.svg)](https://godoc.org/github.com/andygrunwald/perseus)

Local git mirror for your PHP ([composer](https://getcomposer.org/)) project dependencies that works together with [Satis](https://github.com/composer/satis).

*perseus* is a successor out of and drop-in replacement for [Medusa](https://github.com/instaclick/medusa).

## Table of contents

- [Whats wrong with Medusa?](#whats-wrong-with-medusa)
- [Features](#features)
- [Installation](#installation)
	- [From binary](#from-binary)
	- [From docker image](#from-docker-image)
	- [From source](#from-source)
- [Usage](#usage)
	- [Add a new package](#add-a-new-package)
	- [Mirror all packages](#mirror-all-packages)
	- [Update all mirrored packages](#update-all-mirrored-packages)
- [Configuration](#configuration)
	- [Command line flags](#command-line-flags)
	- [`medusa.json` configuration file](#medusajson-configuration-file)
- [Drop-in replacement](#drop-in-replacement)
- [Development](#development)
	- [Build](#build)
	- [Build the docker image](#build-the-docker-image)
	- [Unit tests](#unit-tests)
	- [Release a new version](#release-a-new-version)
- [Project background](#project-background)
	- [The name "*perseus*"](#the-name-perseus)
- [Credits](#credits)

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

### From binary

We offer pre-compiled binaries of *perseus* for easy usage.

1. Checkout our [Releases page](https://github.com/andygrunwald/perseus/releases) and choose the latest release
2. Select your system architecture and operating system of your choice
3. Download and extract the archive
4. Switch to the extracted directory and fire a `perseus version`
5. Have fun

### From docker image

*perseus* is available as Docker image at [andygrunwald/perseus](https://hub.docker.com/r/andygrunwald/perseus/). To download the image, fire:

```sh
$ docker pull andygrunwald/perseus
```

Commands can be executed like a normal installation in the format:

```sh
$ docker run andygrunwald/perseus <Command-Name> [Flags] <Parameter>
```

E.g. the `add` command:

```sh
$ docker run andygrunwald/perseus add --with-deps symfony/console /var/config/medusa-small.json
```

Inside the container, example *medusa* and *satis* configurations from the [.docker](./.docker/) folder are available in the path `/var/config`.
Those can be used to play around.

### From source

To install *perseus* from source, a running [Golang installation](https://golang.org/doc/install) is required.

```sh
$ go get github.com/andygrunwald/perseus
$ cd $GOPATH/src/github.com/andygrunwald/perseus
$ go get ./...
$ make install
$ $GOPATH/bin/perseus
```

## Usage

### Add a new package

The `add` command will mirror the given *<Package-Name>* down to disk (with dependencies if requested) and adds the package into the configured Satis.json file.

Usage:

```sh
$ perseus add <Package-Name> [Config-File]
```

Examples:

```sh
$ perseus add "twig/twig"
$ perseus add --with-deps "symfony/console"
$ perseus add --with-deps "guzzlehttp/guzzle" /var/config/medusa.json
```

### Mirror all packages

The `mirror` command will mirror all configured packages from `medusa.json` down to disk (incl. dependencies) and adds all packages into the configured `satis.json` file.

Usage:

```sh
$ perseus mirror [Config-File]
```

Examples:

```sh
$ perseus mirror
$ perseus mirror /var/config/medusa.json
```

### Update all mirrored packages

The `update` command will update all mirrored packages that are located at disk and update them to the latest state. To find all packages it will do a search in the path configured at `repodir`.

Usage:

```sh
$ perseus update [Config-File]
```

Examples:

```sh
$ perseus update
$ perseus update /var/config/medusa.json
```

## Configuration

*perseus* has two different kinds of configurations:

1. Process settings (via command line flags)
2. `medusa.json` configuration file

### Command line flags

Several settings can be set by command line flags:

* Flag `--config`: Path to the *medusa.json* configuration (default: `medusa.json`)
* Flag `--numOfWorkers`: Number of worker used, when a concurrent process is started (default: number of available CPUs)

### `medusa.json` configuration file

*Perseus* is mainly configured with a JSON file (like Medusa).
Here is a minimalistic example:

```json
{
    "repositories": [
        {
            "name": "myvendor/package",
            "url": "git@othervcs:myvendor/package.git"
        },
        ...
    ],
    "require": [
        "symfony/symfony",
        "monolog/monolog",
        ...
    ],
    "repodir": "/tmp/perseus/git-mirror",
    "satisurl": "http://php.pkg.company.tld/git-mirror",
    "satisconfig": "./satis.json"
}
```

In the next sections an explaination of the single configuration parts can be found.

#### `repositories`

A list of custom packages that are not available on the configured https://packagist.org/.
Per each repository, a name and a url must be given.

#### `require`

A list of repositories to mirror down to disk.

The packages will be searched on the given Packagist instance.
Per default the standard instance https://packagist.org/ will be used.

#### `repodir`

Directory to write all repositories to.

This directory needs to be writable.

#### `satisurl`

URL of the future satis installation.

This URL will be used to prefix all package URLs in the final satis configuration.

#### `satisconfig`

At the end of the run, *perseus* write a valid [satis](https://getcomposer.org/doc/articles/handling-private-packages-with-satis.md#satis) configuration file.
In this setting a valid path to a writeable satis configuration is expected.
Further more the file needs to be exists before and it needs to be a valid satis configuration.

*preseus* itself will only touch and edit the `repositories` section in this satis configuration.
All other parts of the file will be untouched.

##### Example `satis.json`

```json
{
    "archive": {
        "directory": "dist",
        "format": "tar",
        "prefix-url": "http://php.pkg.company.tld/packages/",
        "skip-dev": true
    },
    "homepage": "http://php.pkg.company.tld/packages/",
    "name": "private php package repositories",
    "providers": true,
    "repositories": [
        {
            "type": "git",
            "url": "http://php.pkg.company.tld/git-mirror/symfony/debug.git"
        },
        ...
    ],
    "require-all": true
}
```

## Drop-in replacement

We are a Drop-in replacement for [Medusa](https://github.com/instaclick/medusa).
We have the same command structure and functionality.

But in one point we are not compatible: Logging.
We log way more information during the process as the original Medusa.

Be aware: If you parse the logs of the original Medusa process, you might have to adjust your scripts.

## Development

### Build

To build the application, fire

```sh
$ make build
```

A binary, called `perseus` should appear in the same directory.

### Build the docker image

To build the docker image on your own machine, fire

```sh
$ docker build -t andygrunwald/perseus .
```

### Unit tests

A running go installation is required to execute unit tests.
To execute them, run:

```sh
$ make test
```

Tip: If you plan to contribute via a Pull Request, the use of unit tests is encouraged.

### Release a new version

We use [goreleaser](https://github.com/goreleaser/goreleaser) to build and ship our releases. Further more we follow [semantic versioning](http://semver.org/).
Here is how you create and ship a ner version:

1. Export a `GITHUB_TOKEN` environment variable with the `repo` scope selected. This will be used to deploy releases to your GitHub repository. Create yours [here](https://github.com/settings/tokens/new).

	```sh
	$ export GITHUB_TOKEN="YOUR_TOKEN"
	```

2. GoReleaser uses the latest [Git tag](https://git-scm.com/book/en/v2/Git-Basics-Tagging) of your repository.
Create a tag:

	```console
	$ git tag -a v0.1.0 -m "First release"
	```

	If you don't want to create a tag yet but instead simply create a package based on the latest commit, then you can also use the `--snapshot` flag.

3. Now you can run GoReleaser at the root of your repository:

	```sh
	$ goreleaser
	```

4. That's it! Check your GitHub project's release page.

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
