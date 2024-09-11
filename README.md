# cloudy-files

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/b9540be307fd4fbbbd9fa2018ba43ff5)](https://app.codacy.com/gh/peteraba/cloudy-files/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade)
[![Coverage Status](https://coveralls.io/repos/github/peteraba/cloudy-files/badge.svg?branch=main)](https://coveralls.io/github/peteraba/cloudy-files?branch=main)

Cheap, cloud-based file management

At this point it's considered to be an experiment to have a simple file sharing system with some authorization mechanism
without the need for a database. It's potentially a good fit for small companies or personal use. If you anticipate
more than a few users, you should definitely consider a more robust solution.

## Security

Please note that security for this application is not a top priority. Only use it in production if you know what you
are doing. It is probably still fine for personal or low-scale use, best practices are followed in general.

To be even more specific, the application uses secure cookies to transfer session data. It is encrypted, but
authorization is still travelling through the wire, instead of being looked up on the server side. This means that in
case someone manages to steal your cookie, they can potentially impersonate you.

This decision was deliberate to keep the application simple and to avoid the need for a database.

## TODO

- [ ] Add missing HTML endpoints
- [ ] CSRF protection for all POST/PUT/DELETE web requests
- [ ] Bearer token protection for API requests
- [ ] Test API server
- [ ] Test web server
- [ ] Add lambda support
