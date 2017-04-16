# Surfboard Exporter

Arris Surfboard signal metrics exporter for the [Prometheus](https://prometheus.io)
monitoring system.

## Compatibility

- Arris Surfboard SB6190

## Installing

### Precompiled binaries

Precompiled binaries are available on the [releases page](https://github.com/ipstatic/surfboard_exporter/releases).

### Docker image

Docker images are available on [Docker Hub](https://hub.docker.com/r/ipstatic/surfboard_exporter).

You can launch the container like so:
```bash
docker run --name surfboard_exporter -d -p 9239:9239 ipstatic/surfboard_exporter:2.0.0
```

### Building from source

To build Prometheus from the source code yourself you need to have a working
Go environment with [version 1.5 or greater installed](http://golang.org/doc/install).

## License

MIT, see [LICENSE](https://github.com/ipstatic/surfboard_exporter/blob/master/LICENSE).
