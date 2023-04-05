# GoZix OpenTelemetry

## Dependencies

* [viper](https://github.com/gozix/viper)
* [zap](https://github.com/gozix/zap)

## Configuration

```json
{
  "otel": {
    "connection_type": "collector",
    "collector": {
      "endpoint": "http://jaeger-all-in-one:14268/api/traces"
    }
  }
}
```
