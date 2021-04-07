# JSONPromPoster

This tool was written to upload a JSON metrics file to a Prometheus Collector
endpoint.  As metrics are collected from systems which may be dynamic, be it in
the cloud somewhere, it is important that these endpoints can be scraped by
prometheus.  Hence, this tool will scrape the local c2 metrics and post them to
a central aggregator.  For more details on how to stand up a Prometheus
Collector / PushGateway go here https://github.com/pschou/prom-collector .

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
