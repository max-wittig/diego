# Diego

Simple go script to monitor the status of your containers.
Currently it just outputs to stdout

## Usage

```
Usage of diego:
  -executor string
        Set executor to watch. (default "docker")
  -interval int
        Interval to watch in milliseconds, if watch supplied. (default 1000)
  -output string
        Where to output to. (default "stdout")
  -version
        Print current version
```

### Trivia

* Name was inspired by [@dlouzan](https://github.com/dlouzan), A Coruña

* The project is basically [bernard](https://github.com/max-wittig/bernard)
  for containers.

* `Docker` and `Podman` are supported as container executors
