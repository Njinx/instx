from typing import Coroutine
import utils
from utils import LatencyResponse
import config
import asyncio
import urllib.parse
from common import Canidate, Instance, Instances

def is_outlier(avgs: float, latency: float, weight: float) -> bool:
    if (latency*weight > avgs*config.OUTLIER_MULTIPLIER) or (latency < 0):
        return True
    else:
        return False

def find_canidates(instances: Instances) -> list[Canidate]:
    timing_avgs = instances.get_timing_avgs()
    
    canidates: list[Canidate] = []
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

        canidates.append(Canidate(inst, score))

    canidates.sort(key=lambda c: c.score)

    # Now that we've weeded out the bad instances, lets conduct some actual latency 
    # tests for more accurate results.
    test_results: list[LatencyResponse] = asyncio.run(utils.do_latency_tests(
        list(map(lambda x: urllib.parse.urlparse(x.instance.url).netloc, canidates))))

    asyncio.run(_refine_test_canidates(test_results, canidates))
    
    canidates.reverse() # For use as a stack; best canidates need to be on top
    return canidates

async def _refine_test_canidates(
        test_results: list[LatencyResponse],
        canidates: list[Canidate]):

    tasks: list[Coroutine] = []
    for result in test_results:
        tasks.append(_refine_test_canidates_iter(result))

    for ret in await asyncio.gather(*(tasks)):
        if not ret[0]:
            test_result: LatencyResponse = ret[1]
            canidates[:] = [c for c in canidates if c.instance.url != test_result.addr]

async def _refine_test_canidates_iter(result: LatencyResponse) \
        -> tuple[LatencyResponse, bool]:
    if not result.is_alive:
        res_hostname: str = urllib.parse.urlparse(result.addr).netloc
        intensive_result: LatencyResponse = await utils.do_latency_test_intensive(res_hostname)
        if not intensive_result.is_alive:
            return (result, False)

    return (result, True)
