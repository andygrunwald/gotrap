# gotrap

[![Build Status](https://travis-ci.org/andygrunwald/gotrap.svg?branch=master)](https://travis-ci.org/andygrunwald/gotrap)

[Gerrit](https://code.google.com/p/gerrit/), a code review tool, is often used in bigger projects with self-hosted infrastructure like [TYPO3](https://review.typo3.org/), [Android](https://android-review.googlesource.com/), [HPDD (Intel)](http://review.whamcloud.com/), [Qt](https://codereview.qt-project.org/), [OpenStack](https://review.openstack.org/) or [Golang](https://go-review.googlesource.com/).

Using such a self hosted Git infrastructure resp. Gerrit, there is no built-in solution to benefit from hooks triggered by a Github Pull Request, like the continuous integration service [Travis CI](https://travis-ci.org/) or similar.

**gotrap** is a Gerrit <=> Github <=> Travis CI bridge written in Go.

[![How gotrap works](./docs/how-gotrap-works.png)](#how-does-gotrap-works)

A detailed description about every step can be found in [How gotrap works?](#how-gotrap-works).

PS: You are not limited to use Travis CI. You can use every service that can be triggered by a pull request and reports back to the [commit status api](https://developer.github.com/v3/repos/statuses/) :wink:
Travis CI is only used as an example, because it is one of the most popular.

## Table of contents

1. [Features](#features)
2. [Examples](#examples)
3. [Requirements](#requirements)
4. [Installation](#installation)
5. [Usage](#usage)
6. [Configuration](#configuration)
	1. [gotrap `config.json`](#gotrap-configjson)
		1. [Configuration part `gotrap`](#configuration-part-gotrap)
		2. [Configuration part `github`](#configuration-part-github)
		3. [Configuration part `amqp`](#configuration-part-amqp)
		4. [Configuration part `gerrit`](#configuration-part-gerrit)
	2. [Gerrit plugin `replication`](#gerrit-plugin-replication)
	3. [Gerrit plugin `gerrit-rabbitmq-plugin`](#gerrit-plugin-gerrit-rabbitmq-plugin)
		1. [Exchange](#exchange)
		2. [Queue](#queue)
7. [Source code documentation](#source-code-documentation)
8. [Motivation](#motivation)
9. [Alternative implementations](#alternative-implementations)
	1. [Jenkins](#jenkins)
	2. [Gerrit plugin](#gerrit-plugin)
10. [FAQ](#faq)
	1. [How gotrap works](#how-gotrap-works)
	2. [Why JSON as config file format?](#why-json-as-config-file-format)
	3. [Which AMQP broker are supported?](#which-amqp-broker-are-supported)
	4. [What is about the Github API rate limit?](#what-is-about-the-github-api-rate-limit)
	5. [Can i start multiple Travis CI tests in parallel?](#can-i-start-multiple-travis-ci-tests-in-parallel)
11. [License](#license)
12. [Credits](#credits)

## Features

* Gerrit support
* Github support
* Concurrency (can handle more than one changeset per time)
* Multiple projects / branches support
* Exclude changesets by regular expression
* Templatable comments (Gerrit) and Pull Requests (Github)

## Examples

Here are some examples, how gotrap can look like:

* [[BUGFIX] SelectViewHelper must respect option(Value|Label)Field for arrays](https://review.typo3.org/#/c/36909/) @ TYPO3 Gerrit: 
	* [Github PR](https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests/pull/20)
	* [Travis CI build](https://travis-ci.org/typo3-ci/TYPO3.CMS-pre-merge-tests/builds/50994127)
* [[BUGFIX] Map table names in ext_tables_static+adt.sql in Install Tool](https://review.typo3.org/#/c/36859/) @ TYPO3 Gerrit: 
	* [Github PR](https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests/pull/23)
	* [Travis CI build](https://travis-ci.org/typo3-ci/TYPO3.CMS-pre-merge-tests/builds/50994906)

## Requirements

To run *gotrap*, your Gerrit instance has to fulfill the requirements listed below:

* [Gerrit](https://code.google.com/p/gerrit/) in >= v2.9.0 (tested with v2.9.2 & v2.9.4. May work with a lower version)
* Gerrit plugin [gerrit-rabbitmq-plugin](https://github.com/rinrinne/gerrit-rabbitmq-plugin)
* Gerrit plugin `replication` (shipped with Gerrit)

## Installation

```
$ go get
$ go build .
```

## Usage

Trigger the help with:

```sh
$ gotrap -h
Usage of ./gotrap:
  -config="": Path to configuration file.
  -pidfile="": Write the process id into a given file.
  -version=false: Outputs the version number and exits.
```

`-config` is a required setting.
Without a configuration file, *gotrap* won't start.
Please have a look at the [Configuration](#configuration) chapter, how to configure *gotrap* properly.

`-pidfile` will write the process id of the running *gotrap* process into the given file.
This can be used to monitor *gotrap* via [Nagios](https://www.nagios.org/), [Icinga](https://www.icinga.org/) or something similar.

`--version` will outpout the current version number.

## Configuration

### gotrap `config.json`

The main configuration file is *config.json*.
You can copy the template [config.json.dist](https://github.com/andygrunwald/gotrap/blob/master/config.json.dist) and adjust the defaults according to your environment.
To specify this file as configuration parameter, supply the path to *gotrap* using the `--config` parameter.
The configuration is splitted into several parts.
Below, you will find a description of every part and setting with examples.
Words written in uppercase are "variables" which *have to* be replaced by you.

If you have any question regarding the configuration, please open an issue. We will try to answer it and extend the documentation.

#### Configuration part `gotrap`

```json
"gotrap": {
  "concurrent": 1
}
```

`concurrent` specifies the number of changesets / pull requests which are handled by *gotrap* in parallel.
Please take in mind that this number depends on the [Per Repository Concurrency Setting of Travis CI](http://blog.travis-ci.com/2014-07-18-per-repository-concurrency-setting/).
This is handled by a simple semaphore.

#### Configuration part `github`

```json
"github": {
  "api-token": "GITHUB-API-TOKEN",

  "organisation": "typo3-ci",
  "repository": "TYPO3.CMS-pre-merge-tests",

  "branch-polling-intervall": 15,
  "status-polling-intervall": 30,

  "pull-request": {
    "title": "Gotrap: {{.Change.Subject}}",
    "body": [
      "{{.Change.CommitMessage}}",
      "",
      "------",
      "",
      "Details: {{.Change.URL}}",
      "",
      "------",
      "",
      "This PR was created (automatically) by [gotrap](https://github.com/andygrunwald/gotrap) with :heart: and :beer:"
    ]
  }
},
```

*gotrap* needs to create pull requests at Github to trigger services.
The `github` section contains settings for the connection to Github.

The `api-token` setting will be used to authenticate against Github using [Personal API tokens](https://github.com/blog/1509-personal-api-tokens).
These tokens are bound to a user.
You have to create one in your [personal settings](https://github.com/settings/tokens).

To trigger the actions / hooks (like for running Travis CI), a merge request on Github must be created.
`organisation` and `repository` name the repository, where those pull requests will be created.
The example shows the configuration for [typo3-ci/TYPO3.CMS-pre-merge-tests](https://github.com/typo3-ci/TYPO3.CMS-pre-merge-tests).

Before we can create a merge request in the repository specified in `organisation`/ `repository`, we have to ensure that the changesets, which are created in Gerrit, are replicated to Github.
*gotrap* itself will not replicate any git commits / changesets from Gerrit to Github (as such functionality is already implemented in Gerrit with the `replication` plugin).
Instead, *gotrap* only checks, if the branch is already replicated.
If such a branch is not replicated yet, it will wait for `branch-polling-intervall` seconds, before the next check will be made.
This will be repeated until the branch is replicated.

When the branch is replicated and the merge request is created, the services (like Travis CI) that are configured by the owner of the Github repository, will be triggered by Github automatically.
If all these services are finished with their work, they will report back the results to the [Commit Status API](https://github.com/blog/1227-commit-status-api).
*gotrap* will wait, until this has happened.
`status-polling-intervall` specifies the number of seconds to wait until the next check will be done.

`pull-request` is a multiline field.
This text is used as a template to define the Pull Request.
This multiline field will be joined together with new lines (every line is a new line in the end).
The templating logic is based on the [text/template](http://golang.org/pkg/text/template/) package.
Parts enclosed by *{{...}}* are variables and will be replaced by *gotrap* with respective information.
The data structure [gerrit.Message](http://godoc.org/github.com/andygrunwald/gotrap/gerrit#Message) is available for templating for both parts (`pull-request.title` and `pull-request.body`).

#### Configuration Part `amqp`

*gotrap* receives messages through [AMQP](http://www.amqp.org), a message queing protocol,
The AMQP section contains settings for the connection to the AMQP broker, like [RabbitMQ](http://rabbitmq.com).

```json
"amqp": {
  "host": "mq.typo3.org",
  "port": 5672,
  "username": "AMQP-USERNAME",
  "password": "AMQP-PASSWORD",

  "vhost": "AMQP-VHOST",
  "exchange": "AMQP-EXCHANGE",
  "queue": "AMQP-QUEUE",
  "routing-key": "AMQP-ROUTING-KEY",

  "identifier": "gotrap"
},
```

The settings `host`, `port`, `username`, `password` and `vhost` define the connection to the AMQP broker.
`exchange` and `queue` define, the properties, where the information by Gerrit will be sent to.
If the configured `exchange` and `queue` do not exists in the AMQP broker and if the `username` has rights to create those, *gotrap* will create exchange and queue. Valid attributes are described in the Gerrit plugin *gerrit-rabbitmq-plugin* chapter. If you create `exchange` and `queue` in advance on the broker, those have to match these attributes.

`routing-key` depends on your AMQP and Gerrit plugin `gerrit-rabbitmq-plugin` configuration. If you don't have a complex exchange <-> queue setup, a blank value is fine.

`identifier` is a string, which assign a name to a client that will receive messages by AMQP.

#### Configuration Part `gerrit`

*gotrap* needs to communicate with a Gerrit instance.

```json
"gerrit": {
  "url": "https://review.typo3.org/",

  "username": "GERRIT-USERNAME",
  "password": "GERRIT-PASSWORD",

  "projects": {
    "Packages/TYPO3.CMS": {
      "master": true,
      "TYPO3_6-2": true
    }
  },

  "exclude-pattern": [
    "^\\[WIP\\].*"
  ],

  "comment": [
    "Github tests: {{ .CombinedStatus.State }}",
    "",
    "Pull request: {{ .PullRequest.HTMLURL }}",
    "",
    "",
    "{{ range $key, $value := .CombinedStatus.Statuses }}",
      "Service: {{ $value.Context }}",
      "Description: {{ $value.Description }}",
      "URL: {{ $value.TargetURL }}",
      "",
    "{{ end }}"
  ]
}
```

The `url` is the scheme + host + port for the Gerrit instance.
`username` and `password` are credentials, which will be used to authentificate against the Gerrit instance.
Keep in mind that the `username` needs access to the projects configured in `projects` using the REST API to

* GET changeset information by REST endpoint `/changes/`
* POST a comment to a changeset by REST endpoint `/changes/`

The `projects` settings whitelists projects handled by *gotrap*.
One Gerrit instance can handle multiple projects.
One project can contain multiple branches.
Sometimes, you want to test only a few projects per Gerrit instance or a few branches per project.
A branch (e.g. *master* or *TYPO3_6-2*) needs `true` as value.
Otherwise the branch is configured, but disabled.

Keep in mind: Every `project` which should be handled by *gotrap* needs to be configured.
If the `project` defines specifies at least one branch, only these will be handeled. Otherwise, all branches will be handled by *gotrap*.

In the `exclude-pattern` array, you can configure regular expressions to exclude changesets of configured `project` / branches.
With `"^\\[WIP\\].*"` you exclude all Changeset which are starts with "[WIP]" (e.g. [WIP] This is my not finished feature).
WIP means *W*ork *I*n *P*rogress.

`comment` is a multiline field.
This text is used to post the results of the Github Pull Request (e.g. Travis CI) back to the Gerrit Changeset.
This multiline field will be joined together with new lines (every line is a new line in the end).
The templating logic is based on the [text/template](http://golang.org/pkg/text/template/) package.
Parts enclosed by *{{...}}* are variables and will be replaced by *gotrap* with respective information.
The data structure [github.PullRequest](http://godoc.org/github.com/andygrunwald/gotrap/github#PullRequest) is available for templating for `comment`.

### Gerrit Plugin `replication`

All changesets (including patchsets) have to be replicated to Github as branches. Otherwise we won't be able to create pull requests.

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

Please install the [gerrit-rabbitmq-plugin](https://github.com/rinrinne/gerrit-rabbitmq-plugin) according to its documentation in order to publish Gerrit's stream events to a message broker like [RabbitMQ](http://www.rabbitmq.com/).

**Attention**: If the exchange and queue already exists, the attributes have to match the ones listed below. If both don't exist, yet, the user needs the rights to declare and bind them.

#### Exchange

Type    | durable | autoDelete | internal | noWait
------- | ------- | ---------- | -------- | ------
fanout  | false   | false      | false    | false

#### Queue

durable | autoDelete | exclusive | noWait
------- | ---------- | --------- | ------
true    | false      | false     | false

## Source Code Documentation

The source code itself is documented with [godoc](http://godoc.org/golang.org/x/tools/cmd/godoc) according to their [standards](http://blog.golang.org/godoc-documenting-go-code).
You can see it at [gotrap @ godoc](https://godoc.org/github.com/andygrunwald/gotrap).

## Motivation

I was active in the TYPO3 community some time ago.
Most of this time, I was focusing on quality, testing, stability, (custom) tools and similar.
Since TYPO3 was using Gerrit and Travis CI came up.
For all other projects, I started to love Travis CI and I thought it would be cool to get Travis CI-Support for Gerrit changesets with a self-hosted Git infrastructure.

[Steffen Gebert](https://github.com/StephenKing) and me started to talk about this feature. And he liked this idea. Short after this chat I started hacking on this feature. This implementation was the most hackiest PHP code i ever wrote. This code never goes online.

In February 2015 I met Steffen again at [Config Management Camp in Gent, Belgium](http://cfgmgmtcamp.eu/). We talked about this feature again and I started hacking. Again. But this time I wanted to learn a new language and was fascinated by [the go programming language](http://golang.org/).

And here you see the result.

## Alternative Implementations

*gotrap* is maybe not the best solution for this job, but coding this was fun anyway. Have a look below for alternative / possible solution for this problem I can think about.

PS: If you had created such an alternative or know a different way how to solve this problem, let me know. I will be happy to include your way :wink:

### Jenkins

[Jenkins](http://jenkins-ci.org/) is an awesome tool to execute such work as well. A [Gerrit Trigger](https://wiki.jenkins-ci.org/display/JENKINS/Gerrit+Trigger) plugin already exists and works like a charm in several environments.

With the help of Jenkins you can do the same communication like *gotrap*. 
One benefit over *gotrap* would be the log of the single actions / commands will be public visible. 
Maybe helpful to get a better understanding of what is going on.
With Jenkins, you are not limited to Travis CI tests. You can add your tests as you want.
The disadvantage is: You have to host and maintain a Jenkins environment on your own.

Cool possibilities of Jenkins include:

* [Change microversion header name](https://review.openstack.org/#/c/155611/) @ OpenStack Gerrit
* [Add check for non-existing table internal name for delete table](https://review.openstack.org/#/c/156806/) @ OpenStack Gerrit

### Gerrit Plugin

Gerrit supports plugins written in Java.
To run *gotrap*, we require two of them: `gerrit-rabbitmq-plugin` + `replication`.

To transfer this logic to a Gerrit plugin would make sense.
With this, we would not depend on two plugins and the integration with yet another custom tool written in Go.
You don't have to deal with a deployment, configuration and monitoring of *gotrap*.
All configuration can be embedded into Gerrit.

One disadvantages of this will be that you have to keep your plugin in sync with the development of Gerrit (if they change the plugin API).
*gotrap* communicates via their public stream API (which won't be changed hopefully).

## FAQ

### How gotrap Works

[![How gotrap works](./docs/how-gotrap-works.png)](#how-gotrap-works)

1. A contributer pushes a new changeset or patchset to Gerrit.
2. The next two steps will be (nearly) done at the same time
	1. The Gerrit plugin `gerrit-rabbitmq-plugin` will push a new event into the configured RabbitMQ broker.
	2. The Gerrit plugin `replication` will synchronize the new changeset or patchset to the configured Github repository.
3. *gotrap* will receive the notification through the message queue.
4. *gotrap* will check, if the patchset mentined in the notification is the current patchset of the changeset in Gerrit (sometimes contributors push new patchsets really fast and before we started working on the first message. To avoid "double work", we include this sanity check).
5. *gotrap* checks, if the patchset is already synced as branch and creates a new merge request.
6. Github will trigger Travis CI.
7. If Travis CI is finished, it will report back the results to the Commit Status API. 
8. Until Travis CI is done, *gotrap* will check (via long polling) if Travis CI reported the results already.
9. *gotrap* posts the results of the Commit Status API as comment in the changeset of Gerrit and closes the pull request on Github.

### Why JSON as Config File Format?

Because JSON parsing is a standard package in golang and build in into the language. See [encoding/json](http://golang.org/pkg/encoding/json/).

### Which AMQP Brokers are Supported?

[RabbitMQ](http://www.rabbitmq.com/) is the only official supported AMQP broker currently.
Very likely, it works well with other AMQP brokers, but this was not tested.

### What About the Github API Rate Limit?

The Github API (in v3) has a [rate limit](https://developer.github.com/v3/#rate-limiting).
Currently (2015-03-13), the following limits apply:

* Authenticated requests: 5000 requests per hour
* Unauthenticated requests: 60 requests per hour

*gotrap* needs authentication at Github to create pull requests anyways.
So we got 5.000 req / hour. Lets do a small calculation, what this means:

Imagine it takes
* 1 minute to synchronize your patchset to Github
* 0 seconds until your Travis CI tests will start after a merge request
* 30 seconds to execute your tests on Travis CI
* 0 seconds to notify *gotrap* by AMQP message about a new patchset

We configure `branch-polling-intervall` to `15` seconds and `status-polling-intervall` to `10` seconds.

The following API requests will be done until a result is pushed to Gerrit:

* 4 requests to check if the branch is already synced
* 1 request to create the merge request
* 3 requests to check if Travis CI is already finished
* 1 request to add a "closing comment" to the Github merge request
* 1 request to close the merge request

In theory, we can thus handle 5000 / 10 = **500 patchsets per hour**.
Please keep in mind that some requests go wrong or some actions took longer than expected (e.g. scheduling and starting your tests on Travis CI).
So plan some "spare" requests in (production can be hard).

### Can I Start Multiple Travis CI Tests in Parallel?

Yes, you can.
You need to raise the `Concurrent jobs` setting at Travis CI.
See [Per Repository Concurrency Setting](http://blog.travis-ci.com/2014-07-18-per-repository-concurrency-setting/) at the Travis CI blog.

## License

This project is released under the terms of the [MIT license](http://en.wikipedia.org/wiki/MIT_License).

## Credits

* [Wilson Joseph](https://thenounproject.com/wilsonjoseph/) for his [User-Icon from The Noun Project](https://thenounproject.com/search/?q=developer&i=27713) used in the [How gotrap works](#how-gotrap-works) image
