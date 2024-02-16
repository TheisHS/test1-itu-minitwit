# -*- mode: ruby -*-
# vi: set ft=ruby :

unless File.exist? "/etc/vbox/networks.conf"
  # See https://www.virtualbox.org/manual/ch06.html#network_hostonly
  puts "Adding network configuration for VirtualBox."
  puts "You will need to enter your root password..."
  system("sudo bash vbox-network.sh")
end

# The "2" in the first line above represents the version of the configuration object/block below. 2 is the latest version.
Vagrant.configure("2") do |config|

  # The box is the base image that Vagrant will use to create the VM (we are using Ubuntu 20.04 LTS)
  config.vm.box = "generic/ubuntu2204"
  config_ssh_private_key_path = "~\\\.ssh\\\id_rsa"

  # This will create a private network with a DHCP server
  #config.vm.network "private_network", type: "dhcp"

# For two way synchronization you might want to try `type: "virtualbox"`
# The synced_folder will sync the current directory with the /vagrant directory in the VM
# As our source code is in the ./src directory, we will sync that directory with the /vagrant directory in the VM
  config.vm.synced_folder "./src", "/vagrant", type: "rsync"

# The following block will create a VM with the name "deployment_server"
# primary means that this VM will be the first to be started when you run `vagrant up` - we use SQLITE which is already synced from src, so we dont have to worry yet.
  config.vm.define "deploymentserver", primary: true do |server|
          server.vm.provider :digital_ocean do |provider|
            #I need to provide the path to the private key that will be used to connect to the droplet
            #provider.ssh_key_path = "~\\\.ssh\\\id_rsa"
            #provider.ssh_key_name = "Carl Home Machine"
            provider.ssh_key_path = "~\\\.ssh\\\id_rsa"
            provider.token = "shhhh"
            puts "Using Digital Ocean token: #{provider.token}"
            puts "Using ssh_key_path: #{provider.ssh_key_path}"
            provider.image = 'ubuntu-22-04-x64'
            provider.region = 'fra1'
            # this is the smallest droplet size we can use
            provider.size = 's-1vcpu-1gb'
            provider.privatenetworking = true
          end
    server.vm.hostname = "deploymentserver"

    config.vm.provision "shell", privileged: false, inline: <<-SHELL
        echo "Updating and upgrading ubuntu2204..."
        sudo apt-get update
        sudo apt-get -y upgrade
        echo "Downloading go.mod dependencies..."
        wget https://go.dev/dl/go1.22.0.src.tar.gz
        sudo tar -C /usr/local -xvf go1.12.6.linux-amd64.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        go mod download
        go mod verify


        # Now we will run the minitwit application on the deployment_server
        nohup go minitwit.go > out.log &
              echo "================================================================="
              echo "=                            DONE                               ="
              echo "================================================================="
              #echo "Navigate in your browser to:"
              #THIS_IP=`hostname -I | cut -d" " -f1`
              #echo "http://${THIS_IP}:5000"
    SHELL
  end
end