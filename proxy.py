import os
import sys
import requests
import proxy
import ipaddress
import config
import pickle
from common import Canidate, Instance, Timings
import common
import updater
from http.server import BaseHTTPRequestHandler, HTTPServer, SimpleHTTPServer


SERVER_LIST: list[Canidate] = []

class Proxy(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def inject_auth(self, headers):
        headers["Authorizaion"] = "Bearer secret"
        return headers

    def parse_headers(self):
        req_header = {}
        for line in self.headers:
            line_parts = [o.strip() for o in line.split(":", 1)]
            if len(line_parts) == 2:
                req_header[line_parts[0]] = line_parts[1]
        return self.inject_auth(req_header)

    def do_GET(self, body=True):
        global SERVER_LIST

        try:
            hostname = SERVER_LIST[0].instance.url
            url = "{}{}".format(hostname, self.path)
            req_header = self.parse_headers()

            # Call the target service
            resp = requests.get(url, headers=req_header, verify=False)

            # Respond with the requested data
            self.send_response(resp.status_code)
            for header in resp.headers:
                self.send_header(header.)
            self.wfile.write(resp.content)

            return
        finally:
            self.finish()

def start():
    #with proxy.Proxy(
    #        port = config.PORT,
    #        hostname = ipaddress.IPv4Address("127.0.0.1")
    #    ) as p:
        

    #    proxy.sleep_loop()

    httpd = HTTPServer(("127.0.0.1", config.PORT), Proxy)
    httpd.serve_forever()

def get_servers() -> list[Canidate]:
    if not os.access(updater.SERVER_PICKLE, os.F_OK):
        print(f"Warning: No servers found, using DEFAULT_INSTANCE \
                ({config.DEFAULT_INSTANCE}). Perhaps updater.py isn't running?", 
            file=sys.stderr)

        return [Canidate(
            common.Instance(
                config.DEFAULT_INSTANCE, common.Timings(0, 0, 0, 0)), 0)]

    with open(updater.SERVER_PICKLE, "rb") as fp:
        return pickle.load(fp)

def main():
    global SERVER_LIST
    SERVER_LIST = get_servers()

    start()


if __name__ == "__main__":
    main()
