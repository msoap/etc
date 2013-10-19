#!/usr/bin/env python

import BaseHTTPServer

class HttpSimpleServer(BaseHTTPServer.BaseHTTPRequestHandler):
    def do_GET(s):
        s.send_response(200)
        s.send_header("Content-type", "text/plain")
        s.end_headers()
        s.wfile.write("Hello from Python")

if __name__ == '__main__':
    server_class = BaseHTTPServer.HTTPServer
    httpd = server_class(('', 8080), HttpSimpleServer)
    httpd.serve_forever()
    httpd.server_close()
