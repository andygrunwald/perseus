![perseus logo](assets/perseus_logo.png)

--------------------------------------------

## Benchmark

### Goals
 
To compare *perseus* with [Medusa](https://github.com/instaclick/medusa), we tested it with a simple benchmark.
This benchmark had two goals:

1. Prove feature / Drop-in replacement parity
2. Compary performance of both

### Environment

* Date / Time of benchmark: 2017-09-05 / ~07:30pm
* Versions:
	* perseus: v0.1.0-alpha
	* Medusa: [fefd033](https://github.com/instaclick/medusa/commit/fefd033c4352e195bfe1e54db24f4c79a9700621) from [instaclick/medusa](https://github.com/instaclick/medusa)
* Configuration files:
	* [medusa-big.json](https://github.com/andygrunwald/perseus/blob/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-big.json)
	* [medusa-medium.json](https://github.com/andygrunwald/perseus/blob/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-medium.json)
	* [medusa-small.json](https://github.com/andygrunwald/perseus/blob/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-small.json)
	* [satis.json](https://github.com/andygrunwald/perseus/blob/49bc60077805b5f08c6a8452f0ffaaca8c760659/.docker/satis.json)
* Test machines:
	* Provider: DigitalOcean
	* Operating System: Ubuntu 16.04.2 x64
	* Machine size: 16 GB RAM / 8 CPUs / 160 GB SSD disk
	* Datacenter region: Amsterdam 2
	* Amount: Two machines (1 x perseus, 1 x medusa)

### Measurement

All measurements are done with linux [time(1)](https://linux.die.net/man/1/time) command.
Check [StackOverflow](https://stackoverflow.com/questions/556405/what-do-real-user-and-sys-mean-in-the-output-of-time1) on how to read and interpret it.

Before every measurement we clear the caches.
For perseus this means:

```sh
$ rm -rf /tmp/perseus
```

For medusa this means:

```sh
$ rm -rf /tmp/perseus
$ rm -rf .cache/composer/
```

After every measurement we

* fire a `cat /var/config/satis.json` to check if all repos were written
* delete `/var/config/satis.json` and restore a backup file for the next measurement

### Installation

Step for step installation for reproducibility.

#### perseus

List of executed commands:

```sh
$ apt-get update && apt-get upgrade
$ wget https://github.com/andygrunwald/perseus/releases/download/v0.1.0-alpha/perseus_Linux_x86_64.tar.gz
$ tar -xvzf perseus_Linux_x86_64.tar.gz
$ ./perseus version
perseus v0.1.0-Alpha-4C8098CE24FA56AC7DFD512EA756F95AD9D941EB linux/amd64 BuildDate: 2017-05-09T16:35:17Z

$ mkdir -p /var/config
$ mv *.json /var/config
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-big.json
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-medium.json
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-small.json
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/49bc60077805b5f08c6a8452f0ffaaca8c760659/.docker/satis.json
$ cp /var/config/satis.json /var/config/satis.json.bkp
```

[Full installation protocol of perseus v0.1.0-alpha](https://gist.github.com/andygrunwald/6cc180c03384920e5f9baa589b311802)

#### Medusa

List of executed commands:

```sh
$ apt-get update && apt-get upgrade
$ apt-get install php7.0 php7.0-curl php7.0-zip
$ php -v
PHP 7.0.15-0ubuntu0.16.04.4 (cli) ( NTS )
Copyright (c) 1997-2017 The PHP Group
Zend Engine v3.0.0, Copyright (c) 1998-2017 Zend Technologies
    with Zend OPcache v7.0.15-0ubuntu0.16.04.4, Copyright (c) 1999-2017, by Zend Technologies

# Install Composer according https://getcomposer.org/download/
$ php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
$ php -r "if (hash_file('SHA384', 'composer-setup.php') === '669656bab3166a7aff8a7506b8cb2d1c292f042046c5a994c43155c0be6190fa0355160742ab2e1c88d40d5be660b410') { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('composer-setup.php'); } echo PHP_EOL;"
$ php composer-setup.php
$ php -r "unlink('composer-setup.php');"
$ mv composer.phar /usr/local/bin/composer

$ git clone https://github.com/instaclick/medusa.git
$ cd medusa/ && composer install
$ bin/medusa
Console Tool

Usage:
  [options] command [arguments]

Options:
  --help           -h Display this help message.
...

$ mkdir -p /var/config
$ mv *.json /var/config
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-big.json
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-medium.json
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/f91672f932fd7f5e4a08de54f7df01fddc863e20/.docker/medusa-small.json
$ wget https://raw.githubusercontent.com/andygrunwald/perseus/49bc60077805b5f08c6a8452f0ffaaca8c760659/.docker/satis.json
$ cp /var/config/satis.json /var/config/satis.json.bkp
```

[Full installation protocol of medusa fefd033 from instaclick/medusa](https://gist.github.com/andygrunwald/6cc180c03384920e5f9baa589b311802)

### Tests

#### `add` command

##### With a small package `symfony/console`

```sh
time ./perseus add --with-deps symfony/console /var/config/medusa-small.json
```

```sh
time bin/medusa add --with-deps symfony/console /var/config/medusa-small.json
```

##### With a bigger package `symfony/symfony`

```sh
time ./perseus add --with-deps symfony/symfony /var/config/medusa-small.json
```

```sh
time bin/medusa add --with-deps symfony/symfony /var/config/medusa-small.json
```

#### `mirror` command

##### With a small configuration file `medusa-small.json`

```sh
time ./perseus mirror /var/config/medusa-small.json
```

```sh
time bin/medusa mirror /var/config/medusa-small.json
```

##### With a medium configuration file `medusa-medium.json`

```sh
time ./perseus mirror /var/config/medusa-medium.json
```

```sh
time bin/medusa mirror /var/config/medusa-medium.json
```

##### With a big configuration file `medusa-big.json`

```sh
time ./perseus mirror /var/config/medusa-big.json
```

```sh
time bin/medusa mirror /var/config/medusa-big.json
```

#### `update` command

##### With a small configuration file `medusa-small.json`

```sh
time ./perseus update /var/config/medusa-small.json
```

```sh
time bin/medusa update /var/config/medusa-small.json
```

##### With a medium configuration file `medusa-medium.json`

```sh
time ./perseus update /var/config/medusa-medium.json
```

```sh
time bin/medusa update /var/config/medusa-medium.json
```

##### With a big configuration file `medusa-big.json`

```sh
time ./perseus update /var/config/medusa-big.json
```

```sh
time bin/medusa update /var/config/medusa-big.json
```