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

  # dependencies to test the package (go, sqlite) and build the debian (ruby, fpm)
  config.vm.provision "dependencies", type: "shell" do |s|
    s.inline = <<-OUT
      apt-get update
      apt-get install -y build-essential golang-1.10 jq make ruby-dev sqlite
      gem install --no-ri --no-doc fpm
    OUT
  end

  config.vm.provision "build-cabby", type: "shell" do |s|
    s.inline = <<-OUT
      export GOPATH=/opt/go
      export PATH=/usr/lib/go-1.10/bin/:$PATH

      SRC_DIR=/opt/go/src/github.com/pladdy/cabby
      mkdir -p $SRC_DIR
      cp -r /vagrant/* $SRC_DIR
      cd $SRC_DIR

      make && make test && make build
      fpm -f \
        -s dir \
        -t deb \
        -n cabby \
        -m "Matt Pladna" \
        --description "A TAXII 2.0 server" \
        --after-install scripts/postinst \
        --deb-user cabby \
        --deb-group cabby \
        --deb-pre-depends make \
        --deb-pre-depends jq \
        --deb-pre-depends sqlite \
        -C build/debian .

      if [ $? -eq 0 ]; then
        cp *.deb /vagrant
      fi
    OUT
  end
end
