import utils
from utils import LatencyResponse
from updater import Instance, Instances
import config
import asyncio

class Canidate():
    def __init__(self, instance: Instance, score: float):
        self.instance = instance
        self.score = score

    def __str__(self) -> str:
        return "[{}] {}".format(self.score, self.instance)

def is_outlier(avgs: float, latency: float, weight: float) -> bool:
    if (latency*weight > avgs*config.OUTLIER_MULTIPLIER) or (latency < 0):
        return True
    else:
        return False

async def find_canidates(instances: Instances) -> list[Canidate]:
    timing_avgs = instances.get_timing_avgs()
    
    canidate_list: list[Canidate] = []
    for inst in instances.instance_list:

        if not isinstance(inst.timings.initial, float) \
        or is_outlier(
                timing_avgs.initial,
                inst.timings.initial,
                config.INITIAL_RESP_WEIGHT):
            continue
        if not isinstance(inst.timings.search, float) \
        or is_outlier(
                timing_avgs.search,
                inst.timings.search,
                config.SEARCH_RESP_WEIGHT):
            continue
        if not isinstance(inst.timings.google, float) \
        or is_outlier(
                timing_avgs.google,
                inst.timings.google,
                config.GOOGLE_SEARCH_RESP_WEIGHT):
            continue
        if not isinstance(inst.timings.wikipedia, float) \
        or is_outlier(
                timing_avgs.wikipedia,
                inst.timings.wikipedia,
                config.WIKIPEDIA_SEARCH_RESP_WEIGHT):
            continue

        score = float(inst.timings.initial)   * 1/(config.INITIAL_RESP_WEIGHT) \
              + float(inst.timings.search)    * 1/(config.SEARCH_RESP_WEIGHT) \
              + float(inst.timings.google)    * 1/(config.GOOGLE_SEARCH_RESP_WEIGHT) \
              + float(inst.timings.wikipedia) * 1/(config.WIKIPEDIA_SEARCH_RESP_WEIGHT)
        score = round(score, 2)

        canidate_list.append(Canidate(inst, score))

    canidate_list.sort(key=lambda c: c.score)

    # Now that we've weeded out the bad instances, lets conduct some actual latency 
    # tests for more accurate results.
    test_results: list[LatencyResponse] = utils.do_latency_tests(
        list(map(lambda x: x.instance.url, canidate_list)))

    await _refine_test_canidates(test_results, canidate_list)
    
    return canidate_list

async def _refine_test_canidates(
        test_results: list[LatencyResponse],
        canidate_list: list[Canidate]):

    [_refine_test_canidates_iter(result, canidate_list) for result in test_results]

async def _refine_test_canidates_iter(result: LatencyResponse, canidate_list: list[Canidate]):
    if not result.is_alive:
        print(f':( no alive {result.addr}')
        intensive_results: LatencyResponse = await utils.do_latency_test_intensive(result.addr)
        if not intensive_results.is_alive:
            canidate_list[:] = [c for c in canidate_list if c.instance.url != result.addr]

