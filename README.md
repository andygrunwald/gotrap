# gotrap

A Gerrit &lt;--> Github &lt;--> TravisCI connection written in Go.

TODO

## Examples

TODO

* https://review.typo3.org/#/c/36909/
* https://review.typo3.org/#/c/36446/
* https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests/pull/22
* https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests/pull/23
* https://review.typo3.org/#/c/36859/
* https://review.typo3.org/#/c/36948/

## Requirements

TODO

* [Gerrit](https://code.google.com/p/gerrit/) in >= v2.9.2 (TODO: Has to be checked)
* [gerrit-rabbitmq-plugin](https://github.com/rinrinne/gerrit-rabbitmq-plugin)
* Gerrit plugin `replication`

## Installation

```
$ go get
$ go build .
```

## Configuration

### gotrap `config.json`

TODO

### Gerrit plugin `replication`

All Changesets (including patchsets) has to be replicated to Github as branches. Otherwise we won`t be able to create pull requests.

Example configuration:
```
[remote "github/TYPO3-ci/TYPO3.CMS-pre-merge-tests"]
  projects = Packages/TYPO3.CMS
  url = https://github.com/TYPO3-ci/TYPO3.CMS-pre-merge-tests.git
  push = +refs/changes/*:refs/heads/changes/*
  authGroup = Git Mirror
  mirror = true
  timeout = 120
```

The most important part of this configuration is the `push` property.
This setting says that `refs/changes/` will be replicated to `refs/heads/changes`.

The Gerrit changeset ref `refs/changes/51/36451/8` will be appear as branch `changes/51/36451/8` on Github.
In the example above in the Github repository [typo3-ci/TYPO3.CMS-pre-merge-tests](https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests).

### Gerrit plugin `gerrit-rabbitmq-plugin`

Please install the `gerrit-rabbitmq-plugin` according their documentation to publish Gerrit`s stream events to a message broker like [RabbitMQ](http://www.rabbitmq.com/).

It is a common pattern to declare the exchange and queue of a AMQP broker. Below the attributes of the exchange and queue are listed.

**Attention**: If the exchange and queue already exists the attributes has to be the same as listed below. If both doesn`t exist yet the user need the rights to declare and bind them.

#### Exchange

Type    | durable | autoDelete | internal | noWait
------- | ------- | ---------- | -------- | ------
fanout  | false   | false      | false    | false

#### Queue

durable | autoDelete | exclusive | noWait
------- | ---------- | --------- | ------
true    | false      | false     | false

## Motivation

## Alternative implementations

TODO

### Jenkins

TODO

### Gerrit plugin

TODO

## FAQ

### Why JSON as config file format?

Because json parsing is a standard package in golang and build in into the language. See [encoding/json](http://golang.org/pkg/encoding/json/).

### Which AMQP broker are supported?

[RabbitMQ](http://www.rabbitmq.com/) is the only official supported AMQP broker currently.
Maybe it works with others as well, but the were not tested.

### What is about the Github API rate limit?

TODO

## License

This project is released under the terms of the [MIT license](http://en.wikipedia.org/wiki/MIT_License).