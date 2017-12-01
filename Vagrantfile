# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"
  config.vm.box_check_update = true
  config.vm.provider "virtualbox" do |v|
    v.cpus = 2
    v.memory = 2048
  end

  # avoid 'Innapropriate ioctl for device' messages
  # see vagrant config doc for more info: https://www.vagrantup.com/docs/vagrantfile/ssh_settings.html
  config.ssh.shell = "bash -c 'BASH_ENV=/etc/profile exec bash'"

  # port forwarding for rethinkdb
  config.vm.network :forwarded_port, guest: 8080, host: 8080   # web ui
  config.vm.network :forwarded_port, guest: 28015, host: 28015 # client
  config.vm.network :forwarded_port, guest: 29015, host: 29015 # cluster

  # run simple inline script
  config.vm.provision "install-rethinkdb", type: "shell", inline: <<-SHELL
    DISTRIB_CODENAME=xenial

    source /etc/lsb-release && echo "deb http://download.rethinkdb.com/apt $DISTRIB_CODENAME main" | sudo tee /etc/apt/sources.list.d/rethinkdb.list
    wget -qO- https://download.rethinkdb.com/apt/pubkey.gpg | sudo apt-key add -

    sudo apt-get update
    sudo apt-get install -y rethinkdb

    sed 's/# bind=127.0.0.1/bind=all/g' /etc/rethinkdb/default.conf.sample > default.conf && sudo mv default.conf /etc/rethinkdb/instances.d/default.conf
    sudo /etc/init.d/rethinkdb restart
  SHELL
end
