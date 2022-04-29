PORT: int = 8080            # Port to run proxy on
UPDATE_INTERVAL: int = 180  # Update interval in minutes
INSTANCES_JSON_URL: str = "https://searx.space/data/instances.json"
DEFAULT_INSTANCE: str = "https://paulgo.io"

INITIAL_RESP_WEIGHT: float = 1.2
SEARCH_RESP_WEIGHT: float = 1.2
GOOGLE_SEARCH_RESP_WEIGHT: float = 0.6
WIKIPEDIA_SEARCH_RESP_WEIGHT: float = 0.8
OUTLIER_MULTIPLIER: float = 2.0
