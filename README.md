# domain_exporter

Exports the expiration time of your domains as prometheus metrics.

#### Environment variables

- `DOMAIN_EXPORTER_URL_PREFIX` — use when HTTP endpoint served with a prefix,
  e.g.: For this endpoint `http://example.org/exporters/domains` set to
  `/exporters/domains`. Not really required since useful only to prevent
  breaking human-oriented links. Defaults to empty string.

## Configuration

On the Prometheus settings, add the `domain_exporter` probe:

```yaml
- job_name: domain
  metrics_path: /probe
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_target
    - target_label: __address__
      replacement: localhost:9222 # domain_exporter address
  static_configs:
    - targets:
      - carlosbecker.com
      - carinebecker.com
      - watchub.pw
```

It works more or less like Prometheus's
[blackbox_exporter](https://github.com/prometheus/blackbox_exporter).

Alerting rules examples can be found on the
[_examples](https://github.com/caarlos0/domain_exporter/tree/main/_examples)
folder.

You can configure `domain_exporter` to always export metrics for specific
domains. Create configuration file (`host` field is optional):

```yaml
domains:
- google.com
- name: reddit.com
  host: whois.godaddy.com
```

And pass file path as argument to `domain_exporter`:

```bash
domain_exporter --config=domains.yaml
```

Notice that if you do that, results are cached, and you should change your job 
`metrics_path` to `/metrics` instead.

## Install

**homebrew**:

```bash
brew install caarlos0/tap/domain_exporter
```

**docker**:

```bash
docker run --rm -p 9222:9222 caarlos0/domain_exporter
```

**apt**:

```bash
echo 'deb [trusted=yes] https://repo.caarlos0.dev/apt/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install domain_exporter
```

**yum**:

```sh
echo '[caarlos0]
name=caarlos0
baseurl=https://repo.caarlos0.dev/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install domain_exporter
```

**deb/rpm/apk**:

Download the `.apk`, `.deb` or `.rpm` from the [releases page][releases] and
install with the appropriate commands.

**manually**:

Download the pre-compiled binaries from the [releases page][releases] or clone
the repository build from source.

[releases]: https://github.com/caarlos0/domain_exporter/releases

## Stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/domain_exporter.svg)](https://starchart.cc/caarlos0/domain_exporter)
