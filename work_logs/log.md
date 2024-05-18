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

## March 1

We fixed a final error with the api, and prepared it for simulation start.

Deciding whether to go directly for a new database setup or look at an ORM for current setup.
We continued with an ORM for now and found inspiration in the link below.

https://blog.logrocket.com/comparing-orm-packages-go/

Issues with data persistency have been fixed, but a few minutes later than start of simulation.

## March 2

Looking into setting up automatic steps on pull requests. This is where unit tests, linting assurances and similar tests will take place.

Making a more effective pipeline: https://circleci.com/blog/reduce-cycle-time-pull-requests/

Added CodeClimate bot to project, this integration can do anything from checking test coverage to linting with a lot of different plugins and settings. For now I have added the "Golint" plugin.
Golint:
- Maintained by the Go developers themselves
- Enforces coding conventions, which hopefully will improve code maintainability. 

The effect of this integration will be seen on pull requests, where the bot automatically will comment on any issues or errors it finds according to the configurations we have set. We can also add a 

## March 8

We have set up Prometheus with Grafana.

We chose these tools as they were directly supported and had substantial documentation online. In addition these were also the tools that were introduced to us in class, and we have not found any arguments for picking other tools.

We have only set up three simple prometheus monitoring values, however our setup makes it easy to extend with further values after deliberating on what we will need to monitor and why.

## March 12

For prometheus and grafana, we had 3 example metrics running for setting up monitoring. Today I've looked into different metrics that could be beneficial to monitor:
- Total requests -> this gives an idea of overall traffic and load. (implemented)
- Response times -> Do we have spikes in response times, or periods of prolonged response times, could indicate performance issues.
- Errors -> Measure errors on our end as well, could indicate or point toward bugs or potential attacks on the system. (implemented)
- Database accesses -> How much do we access the database and can this be optimised? (implemented)
- Memory/CPU Usage -> Overall good idea, but we also have the CPU graph in digital ocean providing similar information.
- User oriented metrics -> Things such as 'new users', 'tweets/hour' etc. Will give us data that could be used to correlate with other areas including performance issues and maybe locating potential bottlenecks.

Current issue is figuring out how to display the correct queries in grafana (Any referring to localhost in the following can be replaced by the production ip for live version). Going to localhost:4000/metrics shows our self-defined metrics as such:

```
...
# HELP test1_prometheus_database_accesses_total Amount of database accesses or operations
# TYPE test1_prometheus_database_accesses_total counter
test1_prometheus_database_accesses_total 12
# HELP test1_prometheus_http_requests_total Number of get requests.
# TYPE test1_prometheus_http_requests_total counter
test1_prometheus_http_requests_total 4
...
```

Going to localhost:9090 opens prometheus interface, but neither ``test1_prometheus_http_requests_total`` nor ``sum(test1_prometheus_http_requests_total)``shows any results and this problem then forwards into grafana on localhost:3000 (they dont show up). The connection from prometheus to grafana should be working (server url is http://prometheus:9090) and it also states that it queries the prometheus api and we can display metrics that come with prometheus inherently, but not find our own.

## March 13

### Automatic Release

Nadja has set up the automatic releases using this guide:
https://circleci.com/blog/publishing-to-github-releases-via-circleci/

The Github access token has been setup by me. The guide said something about a version ... I just set it to always be "automatic".

### Prometheus and Grafana
Continuing with Prometheus and Grafana. Removed the default go-metrics. Tested promauto.NewCounter vs prometheus.NewCounter. Removed the prometheus subspace from our metrics again, to make them more distinct. 

/metrics now only shows our defined ones:
```
# HELP test1_database_accesses_total Amount of database accesses or operations
# TYPE test1_database_accesses_total counter
test1_database_accesses_total 5
# HELP test1_errors_total Amount of errors
# TYPE test1_errors_total counter
test1_errors_total 0
# HELP test1_http_requests_total Number of get requests.
# TYPE test1_http_requests_total counter
test1_http_requests_total 1
```

and going to localhost:9090 and checking the status of current targets, it can clearly find our system and says that the connection is up, but I cannot query any of our own defined metrics, but works perfectly will the default ones. The issue does not seem to persist to Grafana, which should be set up as intended. Current problem is our prometheus cannot find our own metrics, but according to the material given and several online sources and videos, I cannot locate the error. 

### Fix ?

Following this Stack Overflow guide, we found that we could redirect the search to host.docker.internal, and then it found it. Maybe giving the docker container a `container_name` would also have done the job, but we cannot do this for two containers, so we will not do this for now.
https://stackoverflow.com/questions/60679103/cannot-capture-client-metrics-with-prometheus

## March 14

### SQLite to PostgreSQL conversion

We need to convert our SQLite database to a PostgreSQL database. We have chosen to use pgLoader for this task.
The conversion is necessary because SQLite is not suitable for production environments, and we want to use a more robust database system for our application.

As a first step we installed postgresql locally. We then created a new minitwit database with pgAdmin.
pgLoader only works for Linux and MacOS, so we had to use a docker container to run it.
Then we scp'ed a copy of our SQLite database to the container, and ran the conversion by connecting to our new local postgres database. We did have some issues with type casting, 
but with some quick changes in data types as well as updates to schema.sql we could successfully convert.

Obs. for all queries referencing user, we must include "" citation signs, as we can't use user as table name in postgres since it is a reserved keyword which resulted in the database
conversion automatically naming the table "user" (with citation sign).

Testing the application with Postgres database:
Just adding the connectionString to main and running the application was fine. However, the API tests did not pass.
The issue was that the current tests were run on the actual database instead of making a new "testing" database, as we had intended to do... ups.
So to make the tests work, we added a go flag package in order to account for a specified environment, this flag is passed to main from the docker-compose file.
If the environment is test or development, the application will run SQLite, otherwise it will connect to the postgres database.

Missing:
4. Setup maintained database on digitalOcean: Database Cluster

5. Migrate local postgres database to DigitalOcean with pg dump

6. Deploy!!

### SQLite to Managed Database at Digital Ocean SOLUTION
DOWNLOAD DB
scp root@46.101.123.125:~/minitwit/minitwit.db C:\Users\stend\Downloads

MANUALLY CHANGE string to TEXT with db browser
db browser: change tables from string to TEXT

START CONTAINER RUNNING PGLOADER WITH DB FILE MOUNTED
docker run -it --rm --name pgloader -v C:\Users\stend\Downloads\minitwit.db:/data/minitwit.db dimitri/pgloader:latest

USE CONTAINER TO TRANSFER DATA FROM MINITWIT.DB to A PostgreSQL local db called minitwit
pgloader sqlite:///data/minitwit.db pgsql://postgres:{MY_PASSWORD}@172.28.144.1/minitwit

USE PG_DUMP TO MAKE BACKUP FILE WE SEND TO SERVER (LIST OF SQL COMMANDS. --clean COMMANDS FOR DROPPING TABLES)
pg_dump --clean -U postgres -f minitwit_backup minitwit

SYNC WITH DIGITAL OCEAN
psql -d {connection_string} -f C:/Users/stend/Desktop/minitwit_backup

GitHub Accept PR
https://github.com/TheisHS/test1-itu-minitwit/pull/37



## March 22

### Optimizing Query Performance

Today we saw that the load time of the public timeline was quite gradually increasing as the number of database entries increased.
This did of course make sense, however we had thought that limiting the rows fetched from the database would be THE fix. But from further investigation we noted that we of course would still have to order all entries based on the timestamp in order to get the latest posts for the timeline.
We then looked into indexing the timestamp of the messages table along the message_id (which postgres automatically indexes). 
To do this we connected to our managed database on Digital Ocean and ran the following command:
``` CREATE INDEX idx_pub_date ON messages(pub_date); ```
Manually we could see that the load time of the public timeline was significantly reduced from 1.5-3 seconds to around 120-200 ms.
As we were looking at the query statistics we also noted long query time on the queries for unfollowing users and retrieving user id by username:
![img.png](img.png)
So we did a few more indices:
![img_2.png](img_2.png)
And thus gained a significant performance boost on the queries we had issues with:
![img_1.png](img_1.png)
This seems like a significant improvement. However, the new statistics are based on much fewer calls to the database, so we will have to monitor the performance over time to see if the indexing has the desired effect.
and the query performance especially deleting a follower is quite notable:
![img_3.png](img_3.png)
from a mean of 45 ms to 4 ms. We are a bit in doubt of how the indexing overhead affect the query performance. For instance, inserting to the follower table a new following pair is now a bit slower, but seems like a trade-off worth taking at the moment.
We will have to monitor this over time.

## March 28

Updated circle ci pipeline to use hadolint for Dockerfile linting. Two ignore-rules have been added, these are common ignores in online examples. Current threshold is "info" level by default, this means that it will fail if we reach a severity equal to that or higher during linting.

### Dockerfile errors
**api-tests and src-tests:**
- add version to packages
- add --no-cache-dir -> Once a package is installed, it does not need to be re-installed and the Docker cache can be leveraged instead. Since the pip cache makes the images larger and is not needed, it's better to disable it.
- Consolidate run commands -> https://github.com/hadolint/hadolint/wiki/DL3059

### Client/server setup

Setup client-server relationship between our webserver and apiserver. To run the program, it's required that the developer has a minitwit database setup on their computer, and created an `.env` file at the root of the directory with a connection string. The contents of the file will look something like the below, where `<USERNAME>` and `<PASSWORD>` are the local settings for psql on the computer, and `<DBNAME>` is the database name ("Minitwit" on my computer).

`
DATABASE_URL = "postgresql://<USERNAME>:<PASSWORD>@host.docker.internal:5432/<DBNAME>?sslmode=disable"
`

To setup a database with postgres, you can use PGAdmin. Remember to initialise the database by running the sql commands from the `schema.sql` file. 

> NOTE: Instead of the `xx_id integer primary key autoincrement`, you will have to write `xx_id serial primary key`, as autoincrement does not exist in psql.

## April 1

### Loggin with Promtail - Loki - Grafana

Instead of en EFK / ELK / ELFK stack, we have gone with Promtail, Loki and Grafana, as we are already using Grafana in our monitoring setup.

This was a good guide: https://www.youtube.com/watch?v=pnycjg_9M-o
We used this package and setup for promtail client: https://github.com/afiskon/promtail-client/blob/master/examples/client-example/main.go
We used this setup as an example setup for things with the alertmanager: https://github.com/rubencougil/loki

## April 23

### Docker Swarm
Following some API crashes during the weekend, we can see from DigitalOcean that each time the server has crashed the CPU has reached 100%.
As we have not configured docker swarm for our actual deployment server yet, only experimented on a separate cluster of nodes - where it works -, 
we were lead to believe that the load balancer of the routing mesh would enable us to redirect to a “healthy” API replica in case of a crash.
Obs. the reason we had not set up docker swarm for our actual deployment server was that we had some consistency concerns
regarding our logging and monitoring setup across different droplets, and we wanted to make sure that we had a working setup before we actually deployed the swarm cluster, but
this will have to wait until we have a more stable API. As we have configured secrets etc. for the experimental swarm cluster (and use the up-to-date database), the quickest way for us to move on is simply notifying the DevOps team of a new API endpoint.
However, as the inherent “load balancer” of docker swarm is rather low-level, or in other words,
only redirects when a node is “unhealthy” aka. dead, we would like a dedicated load balancer outside the swarm.

### Nginx as a Load Balancer
We have chosen to include a Nginx node as reverse proxy for our services to balance the networking load.
To use Nginx we have to add a new droplet in our DigitalOcean project. Our initial approach was to install docker and create a new image to containerize following this guide: `https://upcloud.com/resources/tutorials/load-balancing-docker-swarm-mode`.
Then, the load balancer would be deployed on its own single-node swarm - this separation would help in keeping the load 
balancers' configuration and operation independent and focus solely on managing traffic. However, for simplicity we ended up running nginx directly on the VM instead of in a container.
To configure nginx seperated the configuration files into two parts, one for the api and one for the webserver.
The configuration file for the api is as follows:
```
upstream api_minitest1_backend {
    server 46.101.135.55:4001;
    server 207.154.235.6:4001;
    server 64.226.85.146:4001;
}

server {
    listen 80;
    listen [::]:80;
    server_name api.minitest1.dk;

    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.minitest1.dk;

    (...) # SSL certificate and key for the domain and other configurations
    }
```
Basically, we have defined an upstream block that includes the IP addresses of the nodes in the swarm cluster, and then we have defined a server block that listens on port 80 and redirects to https.
And then we have a server block that listens on port 443 and includes the SSL certificate and key for the domain (more on this later).
The configuration file for the webserver follows the same pattern, however now we have to include the IP addresses of the webserver nodes in the upstream block instead.

Now, we  have a running nginx node outside our swarm but in the same cloud-environment to handle trafficking and balancing
the load among the active replicas of each service (API and webserver). This does not remove the routing mesh “load balancing” but rather complements its
capabilities to the degree that we expect our API service to continue working even during high traffic. 

Moreover, as from the security topic of last week it is quite important to secure our services with SSL encryption.
Furthermore, as we are about to perform a security assessment of our services, and we need to have a secure connection for Metasploit to see vulnerabilities of our services, this is quite fitting time to add SSL.
Moreover, since we've already made adjustments to the API endpoint for the simulation by migrating to the swarm cluster, it is a good moment to update our domain name and implement SSL security measures as we have to notify the DevOps team of the new endpoint anyway.

### Securing with SSL and adding a domain name
Nginx simplifies the process of obtaining free HTTPS certificates through CertBot.
We followed the instructions provided by CertBot:
`https://certbot.eff.org/instructions?ws=nginx&os=ubuntufocal`
to install CertBot and obtain a certificate for our domain. The domain was provided for free by ITU help desk through one.com, in which 
we create two A records in the domain, one for the API and one for the webserver, pointing to the IP addresses of the droplets:
- Webserver: `https://minitest1.dk`
- API: `https://api.minittest1.dk`

We did run into some issues with the API certificate as the API does not have a website (we think?). After reading
some articles we found this one: `https://www.digitalocean.com/community/tutorials/how-to-acquire-a-let-s-encrypt-certificate-using-dns-validation-with-acme-dns-certbot-on-ubuntu-18-04`
which explains some issues with HTTP validation. 
So, we used the acme-dns-certbot tool which connects CertBot to a third-party DNS service to verify the domain ownership of both our domain names, which worked for us.

