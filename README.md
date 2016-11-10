mackerel-plugin-standard-score
====

Mackerel plugin to create custom metrics standard-score from specified metrics.

## Synopsis

```
$ ./mackerel-plugin-standard-score -h
Usage of ./mackerel-plugin-standard-score:
  -cli-mode
        CLI Mode
  -metric-name string
        Metric Name
  -node string
        Mackerel Node Name (default: use hostname)
  -role string
        Role Name
  -service string
        Service Name
```

## Example of mackerel-agent.conf

```
[plugin.metrics.memory-standard-score]
command = "MACKEREL_APIKEY='XXXX' /path/to/mackerel-plugin-standard-score -service service-name -metric-name memory.used -role role-name -node node-name"
```

## Licence

[MIT](https://github.com/takaishi/mackerel-plugin-standard-score/blob/master/LICENCE)

## Author

[takaishi](https://github.com/takaishi)