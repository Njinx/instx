import icmplib as icmp
import asyncio

# TODO: Refactor this
def json_keys_exists_or_x(obj, x, keys):
    if keys[0] not in obj:
        return x

    if len(keys) == 1:
        if keys[0] in obj:
            return obj[keys[0]]
        else:
            return x
    else:
        return json_keys_exists_or_x(obj[keys[0]], x, keys[1:])

class LatencyResponse():
    def __init__(self, addr: str, avg_latency: float, is_alive: bool, packet_loss: float):
        self.addr = addr
        self.avg_latency = avg_latency
        self.is_alive = is_alive
        self.packet_loss = packet_loss

def _latency_test_repack_results(resp: icmp.Host) -> LatencyResponse:
    return LatencyResponse(
            resp.address,
            resp.avg_rtt,
            resp.is_alive,
            resp.packet_loss)

async def do_latency_tests(urls: list[str]) -> list[LatencyResponse]:
    resp_list: list[icmp.Host] = await icmp.async_multiping(
        urls,
        count=4,
        interval=0.2,
        timeout=1,
        privileged=False)

    ret: list[LatencyResponse] = []
    for resp in resp_list:
        ret.append(_latency_test_repack_results(resp))

    return ret

async def do_latency_test_intensive(url: str) -> LatencyResponse:
    resp: LatencyResponse = await icmp.async_ping(
        url,
        count = 8,
        interval = 2,
        timeout = 4,
        privileged = False)

    return _latency_test_repack_results(resp)
