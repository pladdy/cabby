# configure with version "2"
Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"

  config.vm.box_check_update = true

  config.vm.provider "virtualbox" do |v|
    v.cpus = 2
    v.memory = 2048
  end

  config.vm.network "forwarded_port", guest: 1234, host: 1234, auto_correct: true

  # avoid 'Innapropriate ioctl for device' messages
  # see vagrant config doc for more info: https://www.vagrantup.com/docs/vagrantfile/ssh_settings.html
  config.ssh.shell = "bash -c 'BASH_ENV=/etc/profile exec bash'"

  config.vm.provision "dependencies", type: "shell" do |s|
    s.inline = <<-OUT
      apt-get update
      apt-get install -y golang-1.9 ruby-dev build-essential jq sqlite
      gem install --no-ri --no-doc fpm
    OUT
  end

  config.vm.provision "build-cabby", type: "shell" do |s|
    s.inline = <<-OUT
      export GOPATH=/opt/go
      export PATH=/usr/lib/go-1.9/bin/:$PATH

      mkdir -p /opt/go/src/cabby
      cp -r /vagrant/* /opt/go/src/cabby
      cd /opt/go/src/cabby

      make && make test && make build
      cp build/cabby build/usr/bin
      fpm -f -s dir -t deb -n cabby -d jq -m "Matt Pladna" --description "A TAXII 2.0 server" --after-install build/postinst --deb-user cabby --deb-group cabby -C build/debian .

      if [ $? -eq 0 ]; then
        cp *.deb /vagrant
      fi
    OUT
  end
end
