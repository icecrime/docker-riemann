Docker to Riemann
=================

Simple binary to push Docker events to Riemann.

How does it work
----------------

Events are retrieved from the [`/events`](http://docs.docker.io/reference/api/docker_remote_api_v1.11/#23-misc) Docker API endpoint and used to produce Riemann events such that:

* Host is set to a command line provided id (defaults to the hostname)
* Service is set to the Docker event status ("create", "start", "die", "destroy", ...)
* Description holds the container id
* The event is tagged with "docker"

Because of the hardly numeric nature of these events, they hold a metric value of 1. This can for example allow to use this metric, when fed to Graphite, as an annotation for CPU or memory graphs.

![screenshot_1](https://raw.github.com/icecrime/docker-riemann/master/screenshot/screenshot_1.png)

Usage
-----

    Usage of ./docker-riemann:
      -debug=false: Debug mode: outputs Riemann messages upon sending
      -docker="unix:///var/run/docker.sock": Docker daemon location
      -id="icecrime-air.local": Unique identifier used as Riemann events originating host
      -riemann="tcp://localhost:5555": Riemann service location

