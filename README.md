gostatsd
========

[![Godoc](https://godoc.org/github.com/atlassian/gostatsd?status.svg)](https://godoc.org/github.com/atlassian/gostatsd)
[![Build Status](https://travis-ci.org/atlassian/gostatsd.svg?branch=master)](https://travis-ci.org/atlassian/gostatsd)
[![Coverage Status](https://coveralls.io/repos/github/atlassian/gostatsd/badge.svg?branch=master)](https://coveralls.io/github/atlassian/gostatsd?branch=master)
[![GitHub tag](https://img.shields.io/github/tag/atlassian/gostatsd.svg?maxAge=86400)](https://github.com/atlassian/gostatsd)
[![Docker Pulls](https://img.shields.io/docker/pulls/atlassianlabs/gostatsd.svg)](https://hub.docker.com/r/atlassianlabs/gostatsd/)
[![Docker Stars](https://img.shields.io/docker/stars/atlassianlabs/gostatsd.svg)](https://hub.docker.com/r/atlassianlabs/gostatsd/)
[![ImageLayers Size](https://imagelayers.io/badge/atlassianlabs/gostatsd:latest.svg)](https://imagelayers.io/?images=atlassianlabs/gostatsd:latest)

An implementation of [Etsy's][etsy] [statsd][statsd] in Go,
based on original code from [@kisielk](https://github.com/kisielk/).

The project provides both a server called "gostatsd" which works much like
Etsy's version, but also provides a library for developing customized servers.

Backends are pluggable and only need to support the [backend interface](backend/backend.go).

Being written in Go, it is able to use all cores which makes it easy to scale up the
server based on load. The server can also be run HA and be scaled out, see
[Load balancing and scaling out](https://github.com/atlassian/gostatsd#load-balancing-and-scaling-out).


Building the server
-------------------
From the `gostatsd/` directory run `make build`. The binary will be built in `build/bin/<arch>/gostatsd`.


Running the server
------------------
`gostatsd --help` gives a complete description of available options and their
defaults. You can use `make run` to run the server with just the `stdout` backend
to display info on screen.
You can also run through `docker` by running `make run-docker` which will use `docker-compose`
to run `gostatsd` with a graphite backend.

Configuring the backends
------------------------
Backends are configured using `toml`, `json` or `yaml` configuration file passed through
the `--config-path` flag, see [example/config.toml](example/config.toml).


Sending metrics
---------------
The server listens for UDP packets on the address given by the `--metrics-addr` flag,
aggregates them, then sends them to the backend servers given by the `--backends`
flag (comma separated list of backend names).

Currently supported backends are:

* graphite
* datadog
* statsd
* stdout

The format of each metric is:

    <bucket name>:<value>|<type>\n

* `<bucket name>` is a string like `abc.def.g`, just like a graphite bucket name
* `<value>` is a string representation of a floating point number
* `<type>` is one of `c`, `g`, or `ms` for "counter", "gauge", and "timer"
respectively.

A single packet can contain multiple metrics, each ending with a newline.

Optionally, `gostatsd` supports sample rates and tags (unused):

* `<bucket name>:<value>|c|@<sample rate>\n` where `sample rate` is a float between 0 and 1
* `<bucket name>:<value>|c|@<sample rate>|#<tags>\n` where `tags` is a comma separated list of tags
* or `<bucket name>:<value>|<type>|#<tags>\n` where `tags` is a comma separated list of tags

Tags format is: `simple` or `key:value`.


A simple way to test your installation or send metrics from a script is to use
`echo` and the [netcat][netcat] utility `nc`:

    echo 'abc.def.g:10|c' | nc -w1 -u localhost 8125

Monitoring
----------
Currently you can get some basic idea of the status of the server by visiting the
address given by the `--console-addr` option with your web browser.

Contributing
------------
Contribute more backends by sending pull requests.

Load balancing and scaling out
------------------------------
It is possible to run multiple versions of `gostatsd` behind a load balancer by having them
send their metrics to another `gostatsd` backend which will then send to the final backends.

Using the library
-----------------
In your source code:

    import "github.com/atlassian/gostatsd/statsd"

Documentation can be found via `go doc github.com/atlassian/gostatsd/statsd` or at
https://godoc.org/github.com/atlassian/gostatsd/statsd

[etsy]: https://www.etsy.com
[statsd]: https://www.github.com/etsy/statsd
[netcat]: http://netcat.sourceforge.net/
