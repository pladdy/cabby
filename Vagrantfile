# configure with version "2"
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"

  config.vm.box_check_update = true

  config.vm.provider "virtualbox" do |v|
    v.cpus = 2
    v.memory = 2048
  end

  config.vm.network "forwarded_port", guest: 1234, host: 1234, auto_correct: true # cabby
  config.vm.network "forwarded_port", guest: 8888, host: 8888, auto_correct: true # chronograf

  # avoid 'Innapropriate ioctl for device' messages
  # see vagrant config doc for more info: https://www.vagrantup.com/docs/vagrantfile/ssh_settings.html
  config.ssh.shell = "bash -c 'BASH_ENV=/etc/profile exec bash'"

  # dependencies to test the package (go, sqlite) and build the debian (ruby, fpm)
  config.vm.provision "dependencies", type: "shell" do |s|
    s.inline = <<-OUT
      apt-get update
      apt-get install -y build-essential golang-1.10 jq make ruby-dev sqlite # sendmail
      gem install --no-ri --no-doc fpm
    OUT
  end

  config.vm.provision "tick", type: "shell" do |s|
    s.inline = <<-OUT
      curl -sL https://repos.influxdata.com/influxdb.key | sudo apt-key add -
      source /etc/lsb-release
      echo "deb https://repos.influxdata.com/${DISTRIB_ID,,} ${DISTRIB_CODENAME} stable" | sudo tee /etc/apt/sources.list.d/influxdb.list

      apt-get update
      apt-get install telegraf influxdb chronograf kapacitor
      systemctl enable influxdb && systemctl start influxdb
      systemctl enable kapacitor && systemctl start kapacitor
    OUT
  end

  config.vm.provision "build-cabby", type: "shell", run: "never" do |s|
    s.inline = <<-OUT
      export GOPATH=/opt/go
      export PATH=/usr/lib/go-1.10/bin/:$PATH

      SRC_DIR=/opt/go/src/github.com/pladdy/cabby
      mkdir -p $SRC_DIR
      cp -r /vagrant/* $SRC_DIR
      cd $SRC_DIR

      make cabby.deb

      if [ $? -eq 0 ]; then
        cp *.deb /vagrant
      fi
    OUT
  end
end
