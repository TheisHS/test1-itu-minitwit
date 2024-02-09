# Log

## February 9

We will use the following shebang from now on, as it was recommended.
`#!/usr/bin/env bash`

We set up branch protection of the main branch.

- Require pull requests with approvals: Such that people do not push "wrong" code. Until we have functional status checks, this is necessary.
- Require status checks to pass before merge: Such that people do not push code that breaks our project. We do not have any checks yet.
- Only allow squash merge for pull requests: For a cleaner history in our repository

### Choice of language/technology stack

Crystal with Kemal

#### Feature mapping

|                      | Python3 Flask                                               | Crystal Kemal                                                       | Ruby Sinatra                                                            | Golang                                        |
| -------------------- | ----------------------------------------------------------- | ------------------------------------------------------------------- | ----------------------------------------------------------------------- | --------------------------------------------- |
| Our experience       | Moderate experience                                         | No experience                                                       | No experience                                                           | Moderate experience                           |
| Types                | Dynamically typed                                           | Statically typed                                                    | Dynamically typed                                                       | Statically typed                              |
| Performance          | Moderate performance                                        | High performance                                                    | Moderate performance                                                    | High performance                              |
| SQLite support       | yes                                                         | yes                                                                 | yes                                                                     | yes                                           |
| Middleware support\* | yes                                                         | yes                                                                 | yes                                                                     | yes                                           |
| Release date         | 2010                                                        | 2016                                                                | 2007                                                                    | 2011                                          |
| Deployment           | Deployed using virtual environments to manage dependencies. | Compiled into a single binary executable with all its dependencies. | Requires the presence of the Ruby runtime environment and dependencies. | Can be deployed as single binary executables. |

\* Middleware in web API's is used as a design pattern to intercept and manipulate HTTP requests
https://azure.microsoft.com/en-us/resources/cloud-computing-dictionary/what-is-middleware

#### Benchmark

![Crystal Kemal vs. Ruby Sinatra](./crystal-kemal%20vs%20ruby-sinatra.png)
Crystal Kemal vs. Ruby Sinatra
