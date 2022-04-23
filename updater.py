#!/usr/bin/env python3

import time
import math
import config
import requests
import os
import json
from utils import json_keys_exists_or_x, LatencyResponse
import utils
import asyncio
import the_judge as judge

# "searxng": "required"|"forbidden"|"impartial"
_DEFAULT_CRITERIA: str = """
{
    "minimum_csp_grade": "A",
    "minimum_tls_grade": "A",
    "allowed_http_grades": ["V", "F", "C"],
    "allow_analytics": false,
    "is_onion": false,
    "require_dnssec": true,
    "searxng_preference": "required"
}
"""

_CRITERIA_FPATH = "./criteria.json"

_last_run_time: int = 0

class Criteria():
    def __init__(
            self,
            minimum_csp_grade,
            minimum_tls_grade,
            allowed_http_grades,
            allow_analytics,
            is_onion,
            require_dnssec,
            searxng_preference):
        
        self.minimum_csp_grade = minimum_csp_grade
        self.minimum_tls_grade = minimum_tls_grade
        self.allowed_http_grades = allowed_http_grades
        self.allow_analytics = allow_analytics
        self.is_onion = is_onion
        self.require_dnssec = require_dnssec
        self.searxng_preference = searxng_preference

    @staticmethod
    def from_file(file_path: str):

        if not os.access(file_path, os.F_OK):
            with open(file_path, "w+") as fp:
                fp.write(_DEFAULT_CRITERIA)

        raw_json = None
        with open(file_path, "r") as fp:
            raw_json = json.load(fp)

        minimum_csp_grade = json_keys_exists_or_x(raw_json, "A", ["minimum_csp_grade"])
        minimum_tls_grade = json_keys_exists_or_x(raw_json, "A", ["minimum_tls_grade"])
        allowed_http_grades = json_keys_exists_or_x(
            raw_json, ["V", "F", "C"], ["allowed_http_grades"])
        allow_analytics = json_keys_exists_or_x(raw_json, False, ["allow_analytics"])
        is_onion = json_keys_exists_or_x(raw_json, False, ["is_onion"])
        require_dnssec = json_keys_exists_or_x(raw_json, False, ["require_dnssec"])
        searxng_preference = json_keys_exists_or_x(raw_json, "impartial", ["searxng_preference"])

        return Criteria(
            minimum_csp_grade,
            minimum_tls_grade,
            allowed_http_grades,
            allow_analytics,
            is_onion,
            require_dnssec,
            searxng_preference)

    @staticmethod
    def school_scale_to_int(grade: str) -> int:
        GRADE_CHART = {
            "A+": 100, "A": 95, "A-": 90,
            "B+": 89,  "B": 85, "B-": 80,
            "C+": 79,  "C": 75, "C-": 70,
            "D+": 69,  "D": 65, "D-": 60,
                       "F": 50
        }
        
        if grade in GRADE_CHART:
            return GRADE_CHART[grade]
        else:
            return -100

class Timings():
    def __init__(
            self,
            initial: float|str,
            search: float|str,
            google: float|str,
            wikipedia: float|str):

        self.initial = initial
        self.search = search
        self.google = google
        self.wikipedia = wikipedia

    def __str__(self) -> str:
        return "( I={}, S={}, G={}, W={} )".format(
            self.initial,
            self.search,
            self.google,
            self.wikipedia)

class Instance():
    def __init__(self, url: str, timings: Timings):
        self.url = url
        self.timings = timings

    def __str__(self) -> str:
        return "\"{}\": {}".format(self.url, self.timings)

class Instances():
    _required_criteria = Criteria.from_file(_CRITERIA_FPATH)

    def __init__(self, url: str):
        self.instance_list: list[Instance] = []

        res = requests.get(url)
        json_data = res.json()

        if json_data["instances"] is None:
            return None
        
        for inst_key, inst_val in json_data["instances"].items():
            if "timing" not in inst_val:
                continue

            ## check for criteria
            #if "grade" in inst_val["http"]:
            #    print(inst_val["http"]["grade"], file=sys.stderr)
            csp_grade: str = json_keys_exists_or_x(inst_val, "F", ["http", "grade"])
            tls_grade: str = json_keys_exists_or_x(inst_val, "F", ["tls", "grade"])
            http_grade: str = json_keys_exists_or_x(inst_val, "", ["html", "grade"])
            has_analytics = json_keys_exists_or_x(inst_val, True, ["analytics"])
            is_onion = (json_keys_exists_or_x(inst_val, "", ["network_type"]).lower() == "tor")
            has_dnssec = json_keys_exists_or_x(inst_val, False, ["network", "dnssec"])
            searx_fork = json_keys_exists_or_x(inst_val, "", ["generator"]).lower()

            if Criteria.school_scale_to_int(csp_grade) \
            < Criteria.school_scale_to_int(Instances._required_criteria.minimum_csp_grade):
                #print("{} < {}".format(Criteria.school_scale_to_int(csp_grade), Criteria.school_scale_to_int(Instances._required_criteria.minimum_csp_grade)))
                continue
            if Criteria.school_scale_to_int(tls_grade) \
            < Criteria.school_scale_to_int(Instances._required_criteria.minimum_tls_grade):
                continue
            if http_grade not in Instances._required_criteria.allowed_http_grades:
                continue
            if has_analytics and not Instances._required_criteria.allow_analytics:
                continue
            if is_onion != Instances._required_criteria.is_onion:
                continue
            if not has_dnssec and Instances._required_criteria.require_dnssec:
                continue
            if searx_fork == "searx" \
            and Instances._required_criteria.searxng_preference.lower() == "required":
                continue
            if searx_fork == "searxng" \
            and Instances._required_criteria.searxng_preference.lower() == "forbidden":
                continue

            self.instance_list.append(Instance(
                url = inst_key,
                timings = Timings(
                    initial = json_keys_exists_or_x(
                        inst_val["timing"], -1.0, ["initial", "all", "value"]),
                    search = json_keys_exists_or_x(
                        inst_val["timing"], -1.0, ["search", "all", "median"]),
                    google = json_keys_exists_or_x(
                        inst_val["timing"], -1.0, ["search_go", "all", "median"]),
                    wikipedia = json_keys_exists_or_x(
                        inst_val["timing"], -1.0, ["search_wp", "all", "median"]),
                )
            ))

    # TODO: Remove outliers
    def get_timing_avgs(self):
        avgs = Timings(0.0, 0.0, 0.0, 0.0)
        initial_i: int = 0
        search_i: int = 0
        google_i: int = 0
        wikipedia_i: int = 0

        for inst in self.instance_list:
            if inst.timings.initial > 0:
                avgs.initial += inst.timings.initial
                initial_i += 1
            if inst.timings.search > 0:
                avgs.search += inst.timings.search
                search_i += 1
            if inst.timings.google > 0:
                avgs.google += inst.timings.google
                google_i += 1
            if inst.timings.wikipedia > 0:
                avgs.wikipedia += inst.timings.wikipedia
                wikipedia_i += 1

        avgs.initial /= initial_i
        avgs.search /= search_i
        avgs.google /= google_i
        avgs.wikipedia /= wikipedia_i

        return avgs

    def __str__(self) -> str:
        return "{{\n{}\n}}".format(",\n".join(str(inst) for inst in self.instance_list))

def update_best_server():
    global _last_run_time

    cur_time: int = math.floor(time.time())
    if (cur_time - _last_run_time) / 60 < config.UPDATE_INTERVAL:
        return

    instances = Instances(config.INSTANCES_JSON_URL)
    canidates = asyncio.run(judge.find_canidates(instances))
    [print(v) for v in canidates]

    _last_run_time = math.floor(time.time())


if __name__ == "__main__":
    update_best_server()
