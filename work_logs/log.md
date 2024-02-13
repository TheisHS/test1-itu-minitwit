# Log

## February 9

We will use the following shebang from now on, as it was recommended.
`#!/usr/bin/env bash`

We set up branch protection of the main branch.

- Require pull requests with approvals: Such that people do not push "wrong" code. Until we have functional status checks, this is necessary.
- Require status checks to pass before merge: Such that people do not push code that breaks our project. We do not have any checks yet.
- Only allow squash merge for pull requests: For a cleaner history in our repository

### Feature mapping of ITU Minitwit

- Connect to the database
- Create the database tables
- Query the database
- Get the user id
- Fromat a timestamp for display
- Get a gravatar image for an email
- Before a request, it connects to the DB and gets the user, and after a request it closes the connection to the DB.
- UI:
  - A page showing the public timeline, including all public messages.
  - A page showing the timeline with for the logged in user. If no user is logged in, redirect to the public timeline. The private timeline includes all public messages, and the messages from the people the logged in user follows.
  - A page for user profiles. If the user is followed, display a different message ("You follow this user"/"You don't follow ...")
- Routes:
  - A route to follow other users
  - A route to unfollow other users
  - A route for adding messages
  - A route to register
  - A route to login
  - A route to logout

### Choice of language/technology stack

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

![Crystal Kemal vs. Ruby Sinatra](./crystal-kemal%20vs%20ruby-sinatra.png)
Crystal Kemal vs. Ruby Sinatra

**Decision** We will go with Crystal Kemal.



## February 13

## Containerizing using Docker
We have containerized our application and added following files:
- Dockerfile
- compose.yaml
- .dockerignore (just to keep the filesystem of the container clean...)

First we used:
- $ docker init

'docker init' generates some initial docker-related files like Dockerfile, compose.yaml etc.
However, the Dockerfile included quite a bit of complex and unnecessary image-building instructions, 
making it troublesome to adapt to our project (especially since we're not Docker experts).
Instead, we decided to create the dockerfile manually for clarity - we believe that 
understanding the image-building process from the beginning will pay off in the long run.

Our Dockerfile includes a base-image from https://hub.docker.com/_/golang that defines
necessary dependencies for GoLang to build our image off on.
Further we set up the working directory within the image to include application specific dependencies specified in the go.mod file (so we copied our go.mod file
to the working directory of the image's filesystem in order to download and verify them in our running container). 
The compose.yaml is not really needed yet as we do not require configurations for
additional services. However, we have specified a port and made it ready to use later.
We can now (re)build and run our container from the image specified in the Dockerfile with:
- $ docker compose up --build
