# configure with version "2"
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"

  config.vm.box_check_update = true

  go_version = "1.13"

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
      apt-get install -y --no-install-recommends software-properties-common
      add-apt-repository ppa:longsleep/golang-backports
      apt-get update
      apt-get install -y build-essential golang-#{go_version} jq make ruby-dev sqlite rsyslog
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

      # enable syslog input in telegraf for cabby
      cp /vagrant/vagrant/etc/telegraf/telegraf.d/10-cabby.conf /etc/telegraf/telegraf.d/
      systemctl restart telegraf

      # have rsyslog forward syslogs to telegraf (so it can forward to influxdb)
      cp /vagrant/vagrant/etc/rsyslog.d/50-telegraf.conf /etc/rsyslog.d/
      systemctl restart rsyslog
    OUT
  end

  # nevers; run on request only

  config.vm.provision "build-cabby", type: "shell", run: "never" do |s|
    s.inline = <<-OUT
      export GOPATH=/opt/go
      export PATH=/usr/lib/go-#{go_version}/bin/:$PATH

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

  config.vm.provision "install-cabby", type: "shell", run: "never" do |s|
    s.inline = <<-OUT
      systemctl stop cabby
      dpkg -r cabby
      dpkg -i /vagrant/cabby.deb

      cp /vagrant/vagrant/etc/rsyslog.d/40-cabby.conf /etc/rsyslog.d/
      systemctl restart rsyslog

      /vagrant/vagrant/setup-cabby
      systemctl restart cabby
    OUT
  end

  config.vm.provision "restart", type: "shell", run: "never" do |s|
    s.inline = <<-OUT
      systemctl restart telegraf
      systemctl restart influxdb
      systemctl restart chronograf
      systemctl restart kapacitor
      systemctl restart cabby
    OUT
  end
end
