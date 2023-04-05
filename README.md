# GoZix OpenTelemetry

## Dependencies

* [viper](https://github.com/gozix/viper)
* [zap](https://github.com/gozix/zap)

## Configuration

```json
{
  "telemetry": {
    "connection_type": "collector",
    "collector": {
      "endpoint": "http://jaeger-all-in-one:14268/api/traces"
    }
  }
}
```
