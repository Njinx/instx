# InstX

A client-side service that automatically updates the SearX(NG) instance you use to search based on the best one from [searx.space](https://searx.space). The "best" instance is configuable based criteria such as speed, privacy, and security. It updates the instance by hosting a small web server redirects you to the "best" instance. This web server is then set as your default web browser.

## Installation

### Compiling
1. Build with `go get` and `go build`
2. Set `ExecStart` in `instx.service` to the instx binary and copy it to `~/.config/systemd/`

### Start automatically
1. `systemctl --user daemon-reload`
2. `systemctl --user enable instx`
3. `systemctl --user start instx`

### Set as the default search engine
1. Go to http://localhost:8080/getstarted
2. Right-click on the URL and click "Add InstX"
![Add InstX](./images/getstarted.png)
3. Set "InstX" as the default search engine in your browser
![Set as default search engine](./images/ff_default_search_engine.png)

## Configuration
The default config file is located at `~/.config/instx.yaml` but can be overriden by setting `$SEARX_SPACE_AUTOSELECTOR_CONFIG`.

|Required|YAML Key|Description|Go Data Type|Default Value|
|---|---|---|---|---|
|Yes|default_instance|Fallback instance|string|None|
|Yes|proxy.port|Web server port|int|8080|
|No|proxy.preferences_url|[Apply instance settings automatically](#apply-instance-settings-automatically)|string|None|
|Yes|updater.update_interval|How often all the instances are queried and analyzed (in minutes)|int64|180 (3 hours)|
|Yes|updater.advanced.initial_resp_weight||float64|1.2|
|Yes|updater.advanced.search_resp_weight||float64|1.2|
|Yes|updater.advanced.google_resp_weight||float64|0.6|
|Yes|updater.advanced.wikipedia_resp_weight||float64|0.8|
|Yes|updater.advanced.outlier_multiplier||float64|2.0|
|Yes|updater.criteria.minimum_csp_grade||string|A|
|Yes|updater.criteria.minimum_tls_grade||string|A|
|Yes|updater.criteria.allowed_http_grades||[]string|[V, F, C]|
|Yes|updater.criteria.allow_analytics||bool|no|
|Yes|updater.criteria.is_onion||bool|no|
|Yes|updater.criteria.require_dnssec||bool|no|
|Yes|updater.criteria.searxng_preference||string|required|

## Apply instance settings automatically

Grab the saved preferences url at https://favorite.instance/preferences and paste it in `instx.yaml` in `preferences_url`. No need to cut out the original domain name.

![Instance preferences](./images/preferences_url.png)

