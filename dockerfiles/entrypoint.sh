#!/bin/bash

set -x

echo user: $USER
echo uid: $UID

#ENV USER $USER
#ENV UID $UID



service mysql start

bash -c mysql -u root -p'' < ./migrations/0.sql

# update password
mysql -u root -p'' < ./dockerfiles/entrypoint.sql

$@
