#!/usr/bin/env python
"""
HTTP server for metrics
"""

import traceback
import urlparse
from BaseHTTPServer import BaseHTTPRequestHandler
from BaseHTTPServer import HTTPServer
from collector import SurfboardCollector
from prometheus_client import generate_latest, CONTENT_TYPE_LATEST, REGISTRY
from SocketServer import ForkingMixIn

class ForkingHTTPServer(ForkingMixIn, HTTPServer):
    pass

class SurfboardExporterHandler(BaseHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        BaseHTTPRequestHandler.__init__(self, *args, **kwargs)

    def do_GET(self):
        url = urlparse.urlparse(self.path)
        if url.path == '/metrics':
            try:
                self.send_response(200)
                self.send_header('Content-Type', CONTENT_TYPE_LATEST)
                self.end_headers()
                self.wfile.write(generate_latest(REGISTRY))
            except:
                self.send_response(500)
                self.end_headers()
                self.wfile.write(traceback.format_exc())
        elif url.path == '/':
            self.send_response(200)
            self.end_headers()
            self.wfile.write("""<html><head><title>Surfboard Exporter</title>
            </head><body><h1>Surfboard Exporter</h1><p>Visit <a href="/metrics">
            /metrics</a> to use.</p></body></html>""")
        else:
            self.send_response(404)
            self.end_headers()

def start_http_server(port):
    """
    Run the exporter
    """
    REGISTRY.register(SurfboardCollector())
    handler = lambda *args, **kwargs: SurfboardExporterHandler(*args, **kwargs)
    server = ForkingHTTPServer(('', port), handler)
    server.serve_forever()
