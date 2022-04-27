#!/usr/bin/env python3

import time
import math
import config
import os
import json
from utils import json_keys_exists_or_x
import the_judge as judge
from common import Canidate, Instances
import pickle

SERVER_PICKLE = "./servers.dat"

_last_run_time: int = 0

def update_best_servers():
    global _last_run_time

    cur_time: int = math.floor(time.time())
    if (cur_time - _last_run_time) / 60 < config.UPDATE_INTERVAL:
        return

    instances = Instances(config.INSTANCES_JSON_URL)
    canidates: list[Canidate] = judge.find_canidates(instances)
    
    with open(SERVER_PICKLE, "wb") as fp:
        pickle.dump(canidates, fp)

    _last_run_time = math.floor(time.time())


if __name__ == "__main__":
    update_best_servers()
