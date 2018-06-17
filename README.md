whois2http
==============

[![Software License][ico-license]](LICENSE.md)

A whois to http proxy. It listens for whois connections then makes an http request to the backend server.

Usage
-----

```
Usage of ./whois2http:
  -header value
    	Headers to add to the upstream HTTP request. May be used multiple times.
  -listen string
    	target (default ":43")
  -rate string
    	Rate at which requests can be made in the format <count>-<period> where count is an integer and period is one of S, M, H for second, minute, or hour. (default "3-M")
  -upstream string
    	Upstream to which we should proxy (default "http://example.com:80/whois?format=plain&query={{query}}")
```

Example

     whois2http -listen :43 -header "X-Forwarded-Proto: whois" -rate 10-M -upstream http://mockbin.com/request\?query=\{\{query\}\}

`{{query}}` will be replaced as many times as necessary with the value provided by the client, url encoded.

License
-------

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

[ico-license]: https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square
