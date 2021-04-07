# JSONPromPoster

This tool was written to upload a JSON metrics file to a Prometheus Collector endpoint.

The format of the expected JSON file is as follows:

```
{
"network": "myNetwork",
"observer": "myObserver",
"timestamp": 1617804809,
"host": "myHost",
"key": "systemMon/net/rx",
"@version": "1",
"properties": {
  "engine": "systemMon",
  "host": "myHost",
  "system": "mySystem",
  "units": "raw",
  "site": "mySite"
 },
 "value": 0
}
```

# Example usage:
./JSONPromPoster --post http://10.12.128.249:9550/collector/ data.json
