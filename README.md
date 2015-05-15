# gotrap

[![Build Status](https://travis-ci.org/andygrunwald/gotrap.svg)](https://travis-ci.org/andygrunwald/gotrap)

[Gerrit](https://code.google.com/p/gerrit/), a code review tool, is often used in bigger projects with self hosted infrastructure like [TYPO3](https://review.typo3.org/), [Android](https://android-review.googlesource.com/), [HPDD (Intel)](http://review.whamcloud.com/), [Qt](https://codereview.qt-project.org/), [OpenStack](https://review.openstack.org/) or [Golang](https://go-review.googlesource.com/).
With a self hosted Git infrastructure there is no build in solution to benefit from hooks triggered by a Github Pull Request like the continuous integration service [Travis CI](https://travis-ci.org/) or similar.

**gotrap** is a Gerrit <=> Github <=> TravisCI connection written in Go.

PS: You don`t have to use TravisCI. You can use every service which can be triggered by a pull request and reports back to the [commit status api](https://developer.github.com/v3/repos/statuses/) :wink:
Travis CI is only used as an example, because it is one of the most popular.

## Features

TODO

## Examples

Here are some examples how gotrap can look like:

* [[BUGFIX] SelectViewHelper must respect option(Value|Label)Field for arrays](https://review.typo3.org/#/c/36909/) @ TYPO3 Gerrit: 
	* [Github PR](https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests/pull/20)
	* [Travis CI build](https://travis-ci.org/typo3-ci/TYPO3.CMS-pre-merge-tests/builds/50994127)
* [[BUGFIX] Map table names in ext_tables_static+adt.sql in Install Tool](https://review.typo3.org/#/c/36859/) @ TYPO3 Gerrit: 
	* [Github PR](https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests/pull/23)
	* [Travis CI build](https://travis-ci.org/typo3-ci/TYPO3.CMS-pre-merge-tests/builds/50994906)

## Requirements

To run *gotrap* your Gerrit instance has to fulfil the requirements below, enable and configured two plugins:

* [Gerrit](https://code.google.com/p/gerrit/) in >= v2.9.0 (tested with v2.9.2 & v2.9.4. May work with a lower version)
* Gerrit plugin [gerrit-rabbitmq-plugin](https://github.com/rinrinne/gerrit-rabbitmq-plugin)
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

I was active in the TYPO3 community some time ago.
Most of this time i was focusing on quality, testing, stability, (custom) tools and similar.
Since TYPO3 was using Gerrit and TravisCI came up.
For all other projects i started to love TravisCI and i thought it would be cool to get TravisCI-Support for Gerrit Changesets with a self hosted Git infrastructure.

[Steffen Gebert](https://github.com/StephenKing) and me started to talk about this feature. And he liked this idea. Short after this chat i started hacking on this feature. This implementation was the most hackiest PHP code i ever wrote. This code never goes online.

In February 2015 i met Steffen again at [Config Management Camp in Gent, Belgium](http://cfgmgmtcamp.eu/). We talked about this feature again and i started hacking. Again. But this time i wanted to learn a new language and was fascinated by [the go programming language](http://golang.org/).

And here you see the result.

## Alternative implementations

*gotrap* is maybe not the best solution for this job, but coding this was fun anyway. Have a look below for alternative / possible solution for this problem i can think about.

PS: If you had created such an alternative or know a different way how to solve this problem, let me know. I will be happy to include your way :wink:

### Jenkins

[Jenkins](http://jenkins-ci.org/) is an awesome tool to execute such work as well. A [Gerrit Trigger](https://wiki.jenkins-ci.org/display/JENKINS/Gerrit+Trigger) plugin already exists and works like a charm in several environments.

With the help of jenkins you can do the same communication like *gotrap*. 
One benefit over *gotrap* would be the log of the single actions / commands will be public visible. 
Maybe helpful to get a better understanding of what is going on.
With Jenkins you are not limited to TravisCI tests. You can add your tests as you want.
The disadvantage is: You have to host and maintain a jenkins environment on your own.

But have a look what cool things you can create with jenkins (e.g. DB Datasets CI check):

* [Change microversion header name](https://review.openstack.org/#/c/155611/) @ OpenStack Gerrit
* [Add check for non-existing table internal name for delete table](https://review.openstack.org/#/c/156806/) @ OpenStack Gerrit

### Gerrit plugin

Gerrit support custom plugins written in Java.
To run *gotrap* we require two of them: `gerrit-rabbitmq-plugin` + `replication`.

To transfer this logic to a custom Gerrit plugin would make sense.
With this we don\`t depend on two plugins and a custom tool written in Go.
You don\`t have to deal with a deployment, configuration and monitoring of *gotrap*.
All configuration can be embedded into Gerrit.

One disadvantages of this will be that you have to keep your plugin in sync with the development of Gerrit (if they change the plugin API).
*gotrap* communicates via their public stream API (which won\`t be changed hopefully).

## FAQ

### How does gotrap works?

TODO

### Why JSON as config file format?

Because json parsing is a standard package in golang and build in into the language. See [encoding/json](http://golang.org/pkg/encoding/json/).

### Which AMQP broker are supported?

[RabbitMQ](http://www.rabbitmq.com/) is the only official supported AMQP broker currently.
Maybe it works with others as well, but this was not tested.

### What is about the Github API rate limit?

The Github API (in v3) got a [Rate limit](https://developer.github.com/v3/#rate-limiting).
Currently (2015-03-13) this are

* Authenticated requests: 5000 requests per hour
* Unauthenticated requests: 60 requests per hour

*gotrap* needs an authentication at github to create pull requests.

TODO Talk about timings and reqs per hour.

### Can i start multiple Travis CI tests in parallel?

Yes, you can.
You need to raise the `Concurrent jobs` setting at Travis CI.
See [Per Repository Concurrency Setting](http://blog.travis-ci.com/2014-07-18-per-repository-concurrency-setting/) at the Travis CI blog.

## License

This project is released under the terms of the [MIT license](http://en.wikipedia.org/wiki/MIT_License).