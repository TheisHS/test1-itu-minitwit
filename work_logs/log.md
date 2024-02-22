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

We have decided to change language from Crystal + Kemal to Go. Reason for this is that Crystal and especially Kemal lacks documentation to a point where it becomes almost impossible to work with. The reasoning for Go is that it matches Crystal in the comparison sheet we made, while having a lot more documentation.

Initial refactor (setup connections, middleware and declare endpoints) was done together on one screen. With a setup that worked, we split up the work by dividing the different functionalities between group members. We created a local environment (.env) with a SESSION_KEY that all members of the group should have locally, because we do not want to push any secrets to the repository. We can also mirror the Flask session using a Gorilla library for sessions. Additionally we are using the Gorilla mux library for routing.

We thought of TDD when refactoring to Go and the needed tests.

### Containerizing using Docker

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

### HTML templates

Using these guides:

- https://gowebexamples.com/templates/
- https://www.calhoun.io/intro-to-templates-p3-functions/
- https://www.digitalocean.com/community/tutorials/how-to-use-templates-in-go
- https://pkg.go.dev/html/template
- https://pkg.go.dev/github.com/revel/revel/session

Added UserMessage struct that combines the User and Message object into one pair object, as we need these in a combined list to iterate over.

Import encoding to be able to send requests to gravatar for profile pics.

```
  "crypto/md5"
  "encoding/hex"
```
## February 22
### Automated DigitalOcean VM provision using Vagrant
We have created a Vagrantfile to automate the provisioning of a DigitalOcean remote VM.
We use DigitalOcean as our cloud provider, as we have a $200 credit from GitHub Education. 
Using a cloud provider like DigitalOcean is a good idea, as it allows us to deploy our application to a remote server, and not just run it locally.
This is important, as we want to test our application in a production-like environment with the upcoming simulations. Also, this allows for on-demand vertical scaling, which is a pro as opposed to using some local server.

Our virtualization technique was to use Vagrant.
With Vagrant we can automate the provisioning by providing the build instructions needed to run and deploy our application in the correct environment.
It seems a lot like virtualization with Docker, which is why we simply chose to build our already define Docker container inside the VM/droplet. Specifically, we copy our code
into the vm along our docker files, then we install docker and docker-compose in the VM, and then we build and run our application using the docker-compose.yaml file.
For future use we have discussed pushing our docker image to dockerhub, in which case we can simply pull it from there and run it in the VM, with only our compose.yaml file in the VM.
As opposed to manually creating the droplet, or curling the DigitalOcean API, we can now simply run the command:
```$ vagrant up```. Also, automating this provisioning process ensures that we always have the same environment to work with.

We have added our secrets to as Environment Variables to exclude them from the Vagrantfile and the repository.
Specifically, the API key for DigitalOcean is a crucial secret to keep safe as it allows for the creation of droplets and other resources on DigitalOcean - so hopefully we don't get any unexpected expenses just yet :P 

We also added our newly created API folder (with the API that can handle the requests from the simulation) to the docker-compose.yaml file, so that it is included in the build and run process.

The API will be served on port 5001 and the web application on 5000.
We can now ssh into the VM using the command (granted that the member trying to ssh into the VM has set up an ssh key with DigitalOcean):
```$ ssh root@<ip>```
where ``<ip>`` is the IP of the VM we receive from DigitalOcean when we build our droplet.