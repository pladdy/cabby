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
      apt-get install -y ruby-dev build-essential
      apt-get install -y sqlite
      gem install --no-ri --no-doc fpm
    OUT
  end

  config.vm.provision "install-go", type: "shell" do |s|
    s.inline = <<-OUT
      # install Go
      TMP_ROOT=/vagrant/packages
      mkdir "${TMP_ROOT}"

      SHA256_CHECK_SUM=de874549d9a8d8d8062be05808509c09a88a248e77ec14eb77453530829ac02b
      BASE_URL=https://redirector.gvt1.com/edgedl/go
      GO_PKG=go1.9.2.linux-amd64.tar.gz

      # download package if not already in packages directory
      if [ ! -f "${TMP_ROOT}/${GO_PKG}" ]; then
        echo "Downloading golang ${GO_PKG}..."
        wget -q "${BASE_URL}/${GO_PKG}" -O "${TMP_ROOT}/${GO_PKG}"
      fi

      # get checksum and compare
      DOWNLOAD_SUM=`sha256sum ${TMP_ROOT}/${GO_PKG} | cut -d ' ' -f 1`

      if [ "${SHA256_CHECK_SUM}" = "${DOWNLOAD_SUM}" ]; then
        echo "Installing golang ${GO_PKG}..."
        tar -C /usr/local -xzf "${TMP_ROOT}/${GO_PKG}"
      else
        echo "Checksums don't match: Expected: ${SHA256_CHECK_SUM}, Actual: ${DOWNLOAD_SUM}"
        exit 1
      fi
    OUT
  end

  config.vm.provision "build-cabby", type: "shell" do |s|
    s.inline = <<-OUT
      mkdir -p /opt/go/src/cabby

      export GOPATH=/opt/go
      export PATH=/usr/local/go/bin/:$PATH

      cp -r /vagrant/* /opt/go/src/cabby
      cd /opt/go/src/cabby

      make && make test && make build
      if [ $? -eq 0 ]; then
        mv cabby /vagrant/build/usr/local/bin
        cd /vagrant
        fpm -s dir -t deb -n cabby -C build .
      fi
    OUT
  end
end
