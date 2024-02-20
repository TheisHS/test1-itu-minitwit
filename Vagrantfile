# -*- mode: ruby -*-
# vi: set ft=ruby :

# The "2" in the first line above represents the version of the configuration object/block below. 2 is the latest version.
Vagrant.configure("2") do |config|

  # The box is the base image that Vagrant will use to create the VM
  config.vm.box = 'digital_ocean'
  config.vm.box_url = 'https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box'
  config.ssh.private_key_path = '~/.ssh/id_rsa'

  # The synced_folder will sync the current directory with the /app directory in the VM
  config.vm.synced_folder "./src", "/app", type: "rsync"

  # The following block will create a VM with the name "deployment_server"
  # primary means that this VM will be the first to be started when you run `vagrant up` - we use SQLITE which is already synced from src, so we dont have to worry yet.
  config.vm.define "deploymentserver", primary: true do |server|
          server.vm.provider :digital_ocean do |provider, override|
            provider.ssh_key_name = ENV['SSH_KEY_NAME']
            provider.token = ENV["DIGITAL_OCEAN_TOKEN"]
            provider.image = 'ubuntu-22-04-x64'
            provider.region = 'fra1'
            #this is the smallest droplet size we can use
            provider.size = 's-1vcpu-1gb'
            provider.privatenetworking = true
            #this disables the default synced folder - we use rsync instead and specify the folder above (./src -> /app)
            override.vm.synced_folder ".", "/vagrant", disabled: true
          end
          server.vm.hostname = "deploymentserver"

          config.vm.provision "shell", inline: <<-SHELL
            echo "Updating and upgrading ubuntu2204..."
            sudo apt-get update -y
            sudo apt-get upgrade -y

            echo "Installing docker..."
            sudo apt -y install docker.io
            sudo systemctl start docker
            sudo systemctl enable docker
            sudo curl -L "https://github.com/docker/compose/releases/download/v2.2.3/docker-compose-linux-x86_64" -o /usr/local/bin/docker-compose
            sudo chmod +x /usr/local/bin/docker-compose

            #echo "pulling container from dockerhub..."
            #obs. in the future we will do the following:
            #docker pull ... (pull the image from the registry (docker-hub) st. we only need the docker-compose file on the VM)

            echo "re-building and starting container"
            docker-compose -f /app/compose.yaml up --build
            echo "Navigate in your browser to:"
            THIS_IP=`hostname -I | cut -d" " -f1`
            echo "http://${THIS_IP}:5000"

          SHELL
  end
end