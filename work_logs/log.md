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
`$ vagrant up`. Also, automating this provisioning process ensures that we always have the same environment to work with.

We have added our secrets to as Environment Variables to exclude them from the Vagrantfile and the repository.
Specifically, the API key for DigitalOcean is a crucial secret to keep safe as it allows for the creation of droplets and other resources on DigitalOcean - so hopefully we don't get any unexpected expenses just yet :P

We also added our newly created API folder (with the API that can handle the requests from the simulation) to the docker-compose.yaml file, so that it is included in the build and run process.

The API will be served on port 5001 and the web application on 5000.
We can now ssh into the VM using the command (granted that the member trying to ssh into the VM has set up an ssh key with DigitalOcean):

`$ ssh root@<ip>`

where `<ip>` is the IP of the VM we receive from DigitalOcean when we build our droplet.

## February 25

### CI/CD Service

As our code is stored on Github, we've eliminated Gitlab CI as as an option. Travis CI is only free for a single month (for students), and we've also eliminated this.

In reality, it

| Github Actions                                                                                     | CircleCI                                                                                                                  | Our considerations                                                                                                                                                           |
| -------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Is free ... "Cheapest for people with public repositories"                                         | 3000 minutes for free pr month                                                                                            | We have a public repo                                                                                                                                                        |
| Runs full pipeline automatically                                                                   | Can be paused and wait for human interaction                                                                              | We don't have a usecase for needing human intervention before deploying if the code passes all the tests we stup, and the CircleCI feature (even though nice) is not needed. |
| More than CI/CD - can also automate manual tasks like generating changelogs or versioning releases | Only CI/CD, but specialised in this.                                                                                      | We only need CI/CD for now.                                                                                                                                                  |
| Slower than CircleCI                                                                               | Faster than Github Actions                                                                                                | Do we need speed?                                                                                                                                                            |
| Only Windows, MacOS and Linux                                                                      | Every operating system                                                                                                    | We only need Linux                                                                                                                                                           |
| Configuration can be split in mulitple files                                                       | Single file configuration                                                                                                 | Cleaner setup with GHA?                                                                                                                                                      |
| Docker support is still a bit buggy on GHA, works only with Linux.                                 | CircleCI has perfected its Docker support over the years to make it (almost) the de-facto environment for running builds. | We will use docker.                                                                                                                                                          |
| More granular control by exposing all commands. Complexity increase.                               | Less complex, has built in commands for often-used services. Less control                                                 | We don't know what we need yet - so maybe more control is nice, but it being easy is also nice.                                                                              |

https://coderonfleek.medium.com/circleci-vs-github-actions-a-software-engineers-perspective-14567e539b9c
https://www.techtarget.com/searchsoftwarequality/tip/CircleCI-vs-GitHub-Actions-CI-CD-platform-comparison


### CI/CD with CircleCI

CircleCI for the pipeline: https://app.circleci.com/pipelines/circleci/WCqKgj4HsZ36SUyELrTZdu
Docker images on DockerHub (Naddi's account)
VM on Digital Ocean as usual, port 4000 and 4001.

We had some problems with ssh'ing into the VM, but this article helped:

- https://erenbasaran.medium.com/digitalocean-permission-denied-publickey-solution-6cd963049fce

#### Automatic releases

We will not do this yet. See article below:

- https://circleci.com/blog/publishing-to-github-releases-via-circleci/

```yaml
  release:
    triggers:
      - schedule:
          cron: "0 20 * * 4"
          filters:
            branches:
              only:
                - main
```

`0 20 * * 4` means "in the 0th minute, in the 20th hour, on whatever day of the month, in whatever month, on the 4th day of the week".

## February 26

### Tests as part of our pipeline

Docker compose files should specify that the context is in the root directory but that the docker file is in a subdirectory. This is how the Dockerfiles are set up (for them to work in the pipeline), so compose files are fixed.

Tests are inserted in their own directories for them to have a Dockerfile. Used this page to get the format of the compose.test files, and the command that kills the server container when the tests are done (used in next paragraph).
https://stackoverflow.com/questions/40907954/terminate-docker-compose-when-test-container-finishes

The `run-tests.sh` script builds the two images for testing the services. Then it spins up two separate containers for testing. It exits with an error code if they do not pass, which should make our pipeline fail.

When running the service in one container and the python test scripts in another, I continously got a ConnectionError, which indicated that the two containers did not communicate. To fix this, I've tried many different things, but the fix that did it was make the containers run on localhost. This is as far as I'm aware not best practice, but I could not get it to work in any other way. 
https://stackoverflow.com/questions/43547795/how-to-share-localhost-between-two-different-docker-containers

Dockerfiles now have a "database" argument. If as database is passed (e.g. minitwit.db) this is copied into the image. If no database is given, the empty.db is copied instead and the gofile runs on an empty database. This is for the tests.

A "run_tests" step is added to our pipeline and the build steps are made dependent on this.

TODO: 
1. Don't reset database on deployment. 
2. Maybe parallellise the tests to make pipeline faster.

## February 28

We want to persist our database, so it's not reset every time we rebuild our docker images - this happens on every deploy.
We use volumes to persist the minitwit.db file:
https://www.baeldung.com/ops/docker-mount-single-file-in-volume

## February 29

Docker volumes to solve issues with persistent data.
- Add volumes tag in bottom of compose to define the volume and its configuration
- Then define inside the container which volume that container should have access to and be mounted with.
- We are using the format “named volumes”, <name>:<path_to_db>
- Code is failing because we tested some things with the path to db and this makes the tests fail.
- Best result is being able to create the volumes (check with docker volume ls)
- The db is not persisted after a re-deploy and this is the issue as of now

## February 30

We fixed a final error with the api, and prepared it for simulation start.

Deciding whether to go directly for a new database setup or look at an ORM for current setup.
We continued with an ORM for now and found inspiration in the link below.

https://blog.logrocket.com/comparing-orm-packages-go/

Issues with data persistency have been fixed, but a few minutes later than start of simulation.