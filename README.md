whois2http
==============

[![Software License][ico-license]](LICENSE.md)

A whois to http proxy. It listens for whois connections then makes an http request to the backend server.

Usage
-----

     whois2http -listen :43 -header "X-Forwarded-Proto: whois" -upstream http://example.com/whois?format=plain&query={{query}}

`{{query}}` will be replaced as many times as necessary with the value provided by the client, url encoded.

TODO
----


Future Improvements
-------------------

License
-------

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.

[ico-license]: https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square
