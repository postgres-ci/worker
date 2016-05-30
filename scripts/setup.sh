framework_src=https://raw.githubusercontent.com/postgres-ci/assert/master/framework/assert.sql
framework_dst=/opt/postgres-ci/assets/assert/framework.sql

[ -d /opt/postgres-ci/assets/assert ] || mkdir -p /opt/postgres-ci/assets/assert

wget -O ${framework_dst} ${framework_src}

cp opt.sh  /opt/postgres-ci/assets/setup.sh

chmod +x /opt/postgres-ci/assets/setup.sh