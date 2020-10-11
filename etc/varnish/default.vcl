#
# This is an example VCL file for Varnish.
#
# It does not do anything by default, delegating control to the
# builtin VCL. The builtin VCL is called when there is no explicit
# return statement.
#
# See the VCL chapters in the Users Guide at https://www.varnish-cache.org/docs/
# and https://www.varnish-cache.org/trac/wiki/VCLExamples for more examples.

# Marker to tell the VCL compiler that this VCL has been adapted to the
# new 4.0 format.
vcl 4.0;

# Default backend definition. Set this to point to your content server.
backend default {
    #.host = "127.0.0.1";
    .host = "10.162.76.103";
    .port = "9292";
} 
sub vcl_recv {
    if (req.url ~ "^/initialize") {
       ban("obj.http.url ~ ^/api/audience/dashboard");
    }
    if (req.url ~ "^/registration\?") {
       set req.url = regsub(req.url, "\?.*", "");
    }
    if (req.http.Cookie) {
      unset req.http.Cookie;
    }
    # Happens before we check if we have this in cache already.
    #
    # Typically you clean up the request here, removing cookies you don't need,
    # rewriting the request, etc.
}

sub vcl_backend_response {
    # Happens after we have read the response headers from the backend.
    #
    # Here you clean the response headers, removing silly Set-Cookie headers
    # and other mistakes your backend does.
    if (beresp.status >= 500 && bereq.is_bgfetch) {
      return (abandon);
    }
    if (beresp.status >= 500) {
      set beresp.uncacheable = true;
    }
    if (beresp.http.content-type ~ "text|javascript|protobuf") {
      set beresp.do_gzip = true;
    }
    if (bereq.url ~ "^/api/audience/dashboard") {
      if (beresp.ttl > 0s) {
        set beresp.ttl = beresp.ttl - 0.1s;
        set beresp.grace = 0.1s;
      }
    } else {
      set beresp.ttl = 86400s;
      set beresp.http.cache-control = "public, max-age=86400";
    }
}

sub vcl_deliver {
    # Happens when we have all the pieces we need, and are about to send the
    # response to the client.
    #
    # You can do accounting or modifying the final object here.
}
